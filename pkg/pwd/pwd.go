package pwd

import "golang.org/x/crypto/bcrypt"

// Hash 明文密码 → bcrypt 哈希（注册、初始化 admin 时用）
func Hash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

// Compare 验证明文和哈希是否匹配（登录时用）
func Compare(hash, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
