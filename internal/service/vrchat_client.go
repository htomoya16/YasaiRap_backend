package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
)

// Search All Users から使う最低限の情報
type VRChatUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// WhitelistService から見えるインターフェース
type VRChatClient interface {
	// displayName 完全一致で1件だけ探す。
	// 0件 -> ErrNoExactMatch
	// 複数件 -> ErrMultipleExactMatch
	SearchExactUserByDisplayName(ctx context.Context, displayName string) (*VRChatUser, error)
}

var (
	ErrNoExactMatch       = errors.New("no exact match user found")
	ErrMultipleExactMatch = errors.New("multiple exact match users found")
)

// HTTP 実装。
// 2FA有効アカウントで /auth/user → /auth/twofactorauth/totp/verify → /users を叩く。
type HTTPVRChatClient struct {
	BaseURL string

	Username   string
	Password   string
	TOTPSecret string
	UserAgent  string

	HTTPClient *http.Client

	mu sync.Mutex // ログイン処理の多重実行防止
}

// 環境変数から読み込む想定:
// VRCHAT_USERNAME / VRCHAT_PASSWORD / VRCHAT_TOTP_SECRET
// 戻り値: HTTPVRChatClient
func NewHTTPVRChatClientFromEnv() (*HTTPVRChatClient, error) {
	u := os.Getenv("VRCHAT_USERNAME")
	p := os.Getenv("VRCHAT_PASSWORD")
	secret := os.Getenv("VRCHAT_TOTP_SECRET")
	contact := os.Getenv("YASAIRAP_CONTACT_EMAIL")

	if u == "" || p == "" {
		return nil, fmt.Errorf("VRCHAT_USERNAME or VRCHAT_PASSWORD is empty")
	}
	if secret == "" {
		return nil, fmt.Errorf("VRCHAT_TOTP_SECRET is empty (2FA TOTP secret required)")
	}

	ua := "yasairap-backend/0.1"
	if contact != "" {
		ua = fmt.Sprintf("%s %s", ua, contact)
	}

	// cookie
	jar, _ := cookiejar.New(nil)

	return &HTTPVRChatClient{
		BaseURL:    "https://api.vrchat.cloud/api/1",
		Username:   u,
		Password:   p,
		TOTPSecret: secret,
		UserAgent:  ua,
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
	}, nil
}

// displayName 完全一致で1件だけ返す。
// ここから呼べば裏で勝手にログイン＋2FA＋Search All Users までやる。
func (c *HTTPVRChatClient) SearchExactUserByDisplayName(
	ctx context.Context,
	displayName string,
) (*VRChatUser, error) {
	if displayName == "" {
		return nil, ErrNoExactMatch
	}

	// まずは既存セッションを信じる or ログインする
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, fmt.Errorf("vrchat login failed: %w", err)
	}

	// 1回目の検索
	user, status, err := c.searchOnce(ctx, displayName)
	if err == nil {
		// 見つかったらVRChatUserを返す
		return user, nil
	}

	// 401 以外のエラーはそのまま返す
	if status != http.StatusUnauthorized {
		return nil, err
	}

	// 401 → セッション切れとみなして一度だけ再ログインして再試行
	if err := c.forceReLogin(ctx); err != nil {
		return nil, fmt.Errorf("vrchat re-login failed: %w", err)
	}

	// 2回目の検索
	user, status, err = c.searchOnce(ctx, displayName)
	if err == nil {
		// 見つかったらVRChatUserを返す
		return user, nil
	}

	// もう一度401
	if status == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized after re-login")
	}

	return nil, err
}

// 実際に /users を1回叩いて、
// - ユーザーが見つかれば (*VRChatUser, 200, nil)
// - 401なら (nil, 401, error)
// - その他エラーなら (nil, status, error)
// みたいに返すヘルパー。
func (c *HTTPVRChatClient) searchOnce(
	ctx context.Context,
	displayName string,
) (*VRChatUser, int, error) {
	u, err := url.Parse(c.BaseURL + "/users")
	if err != nil {
		return nil, 0, err
	}
	q := u.Query()
	q.Set("search", displayName)
	q.Set("n", "100")
	u.RawQuery = q.Encode()

	// GETで組み立て
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	// GET respose
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	// 401
	if res.StatusCode == http.StatusUnauthorized {
		return nil, http.StatusUnauthorized, fmt.Errorf("unauthorized")
	}
	// 200
	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, fmt.Errorf("vrchat search users status=%d", res.StatusCode)
	}

	var users []VRChatUser
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, res.StatusCode, err
	}

	// displayName完全一致だけ抽出
	matches := make([]VRChatUser, 0)
	for _, u := range users {
		if u.DisplayName == displayName {
			matches = append(matches, u)
		}
	}

	// displayName完全一致がない場合
	if len(matches) == 0 {
		return nil, res.StatusCode, ErrNoExactMatch
	}
	// displayName完全一致が2個以上
	if len(matches) > 1 {
		return nil, res.StatusCode, ErrMultipleExactMatch
	}

	// displayName完全一致が1個
	return &matches[0], res.StatusCode, nil
}

// ------- 認証・2FA周り -------

// 通常の利用時: 必要ならログイン＋2FA。
// 既に有効セッションがあれば何もしない。
func (c *HTTPVRChatClient) ensureLoggedIn(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// すでにセッションがありそうなら、一旦 /auth/user で確認してもいいし、
	// もっと割り切って「一回ログイン済みなら何もしない」でもよい。

	if c.hasAuthCookie() {
		// ここで何もしない → 既存セッションを信じる
		return nil
	}

	// auth cookie なければログイン＋2FA
	return c.loginWith2FA(ctx)
}

func (c *HTTPVRChatClient) hasAuthCookie() bool {
	if c.HTTPClient.Jar == nil {
		return false
	}
	u, _ := url.Parse(c.BaseURL)
	for _, ck := range c.HTTPClient.Jar.Cookies(u) {
		if ck.Name == "auth" && ck.Value != "" {
			// ほんとは有効期限とか見たいが、最低限「あるかどうか」チェック
			return true
		}
	}
	return false
}

// セッション切れなどで強制ログインし直したいとき
func (c *HTTPVRChatClient) forceReLogin(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loginWith2FA(ctx)
}

// /auth/user → (必要なら) /auth/twofactorauth/totp/verify
func (c *HTTPVRChatClient) loginWith2FA(ctx context.Context) error {
	// 1. /auth/user を Basic 付きで叩く
	endpoint := c.BaseURL + "/auth/user"

	// GETで組み立て
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	// Authorization: Basic base64(urlencode(username):urlencode(password))
	req.Header.Set("Authorization", "Basic "+basicToken(c.Username, c.Password))
	req.Header.Set("User-Agent", c.UserAgent)

	// GET respose
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 401
	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: check username/password or IP restrictions")
	}
	// 200
	if res.StatusCode != http.StatusOK {
		// それ以外
		return fmt.Errorf("login /auth/user failed status=%d", res.StatusCode)
	}

	// 2. レスポンスを読んで 2FA 要求状態かどうか判定
	// VRChat.communityの仕様では、2FA有効アカウントで pending な場合や
	// requiresTwoFactorAuth が返るケースがある。ここではそれを見てTOTPを投げる。:contentReference[oaicite:2]{index=2}
	var cu struct {
		RequiresTwoFactorAuth []string `json:"requiresTwoFactorAuth"`
	}
	if err := json.NewDecoder(res.Body).Decode(&cu); err != nil {
		// レスポンスを使わない運用もあるので、パース失敗しても致命ではない扱いにするならここで return nil でもよい。
		// ただし docs 的には CurrentUser が返る前提なので、一旦エラー扱いにしておく。
		return err
	}

	// requiresTwoFactorAuth に "totp" が含まれていたら TOTP 検証が必要
	needsTOTP := false
	for _, v := range cu.RequiresTwoFactorAuth {
		if v == "totp" {
			needsTOTP = true
			break
		}
	}

	// TOTP検証
	if needsTOTP {
		return c.verifyTOTP(ctx)
	}

	// requiresTwoFactorAuth が空 → 2FA不要 or 既にこのセッションは2FA済み → 何もしない。
	return nil
}

// /auth/twofactorauth/totp/verify に TOTP を投げる
func (c *HTTPVRChatClient) verifyTOTP(ctx context.Context) error {
	// TOTP用のsecretキーが空
	if c.TOTPSecret == "" {
		return errors.New("TOTP secret not configured")
	}

	// secretキーからTOTP生成
	code, err := totp.GenerateCode(c.TOTPSecret, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to generate TOTP: %w", err)
	}

	endpoint := c.BaseURL + "/auth/twofactorauth/totp/verify"

	// {"code":"123456"}
	body := struct {
		Code string `json:"code"`
	}{
		Code: code,
	}
	// body を JSON の []byte に変換する
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// POSTで組み立て
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	// auth cookie は loginWith2FA の /auth/user で Jar に保存されている前提
	// POST response
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 401
	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("2fa totp verify unauthorized (wrong TOTP or missing auth cookie)")
	}
	// 200
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("2fa totp verify failed status=%d", res.StatusCode)
	}

	// 戻り値 verified / enabled を一応見てもいいが、200なら成功とみなす。
	return nil
}

// Basic認証ヘッダ用トークン生成
func basicToken(username, password string) string {
	raw := url.QueryEscape(username) + ":" + url.QueryEscape(password)
	// base64 エンコード
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
