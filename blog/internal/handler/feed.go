package handler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RSS / sitemap 输出的是原始 XML，给阅读器和搜索引擎看，
// 不走统一 {code,message,data} JSON 包装，所以不经过 goctl 生成的 handler，
// 在这里手工注册裸路由（blog.go 里调用 RegisterFeedRoutes）。

// siteBase 从请求推导站点绝对地址（RSS/sitemap 里的链接必须是完整 URL）。
// 反代后 r.Host 仍是用户访问的域名；https 看 X-Forwarded-Proto（nginx 转发时带上）。
func siteBase(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// publishedArticles 取已发布文章，RSS 和 sitemap 共用
func publishedArticles(r *http.Request, svcCtx *svc.ServiceContext, size int) []*model.Article {
	published := int64(1)
	articles, err := svcCtx.ArticleModel.FindPage(r.Context(), model.ArticleListCond{
		Status: &published,
		Size:   size,
	})
	if err != nil {
		return nil
	}
	return articles
}

func siteTitle(r *http.Request, svcCtx *svc.ServiceContext) string {
	title, err := svcCtx.SettingsModel.Get(r.Context(), "siteTitle")
	if err != nil || title == "" {
		return "博客"
	}
	return title
}

// ---------- RSS 2.0 ----------

type rssFeed struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel rssChannel  `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
}

func rssHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := siteBase(r)
		articles := publishedArticles(r, svcCtx, 20) // RSS 只放最近 20 篇

		items := make([]rssItem, 0, len(articles))
		for _, a := range articles {
			link := fmt.Sprintf("%s/articles/%s", base, a.Slug)
			desc := a.Summary
			if desc == "" {
				desc = a.Title
			}
			items = append(items, rssItem{
				Title:       a.Title,
				Link:        link,
				Description: desc,
				PubDate:     a.PublishedAt.Format(time.RFC1123Z), // RSS 规范用 RFC822 格式
				Guid:        link,
			})
		}

		feed := rssFeed{
			Version: "2.0",
			Channel: rssChannel{
				Title:       siteTitle(r, svcCtx),
				Link:        base,
				Description: siteTitle(r, svcCtx),
				Items:       items,
			},
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write([]byte(xml.Header))
		_ = xml.NewEncoder(w).Encode(feed)
	}
}

// ---------- Sitemap ----------

type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	Urls    []sitemapUrl `xml:"url"`
}

type sitemapUrl struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func sitemapHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := siteBase(r)
		articles := publishedArticles(r, svcCtx, 500) // sitemap 尽量全，个人博客 500 篇绰绰有余

		urls := []sitemapUrl{
			{Loc: base + "/", ChangeFreq: "daily", Priority: "1.0"},
			{Loc: base + "/archive", ChangeFreq: "daily", Priority: "0.6"},
			{Loc: base + "/tags", ChangeFreq: "weekly", Priority: "0.6"},
			{Loc: base + "/about", ChangeFreq: "monthly", Priority: "0.4"},
		}
		for _, a := range articles {
			urls = append(urls, sitemapUrl{
				Loc:        fmt.Sprintf("%s/articles/%s", base, a.Slug),
				LastMod:    a.UpdatedAt.Format("2006-01-02"),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			})
		}

		set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", Urls: urls}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write([]byte(xml.Header))
		_ = xml.NewEncoder(w).Encode(set)
	}
}

// RegisterFeedRoutes 手工注册 RSS / sitemap 裸路由
func RegisterFeedRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/v1/rss", Handler: rssHandler(svcCtx)})
	server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/v1/sitemap.xml", Handler: sitemapHandler(svcCtx)})
}
