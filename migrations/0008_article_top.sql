-- 文章置顶：列表排序 is_top 优先于发布时间
ALTER TABLE article ADD COLUMN is_top INTEGER NOT NULL DEFAULT 0;

-- 置顶展示/排序用不上索引（文章量小），先不加
