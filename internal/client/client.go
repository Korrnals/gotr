package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gotr/internal/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiPrefix = "index.php?/api/v2/"

type HTTPClient struct {
	client  *http.Client
	baseURL *url.URL
}

// Скрытая структура с опциями (не экспортируется!)
type options struct {
	insecure            bool
	timeout             time.Duration
	tlsHandshakeTimeout time.Duration
}

// authTransport автоматически добавляет Basic Auth ко всем запросам
type authTransport struct {
	username string
	apiKey   string
	base     http.RoundTripper
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.username, t.apiKey)
	// Можно ещё добавить заголовки:
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// Значение по умолчанию
var defaultOptions = options{
	insecure:            false,
	timeout:             30 * time.Second,
	tlsHandshakeTimeout: 10 * time.Second,
}

// Тип-функция для опций
type ClientOption func(*options)

// Функция-опция WithInsecureSkipVerify -  вкл/выкл проверку tls-сертификата
func WithSkipTlsVerify(insecure bool) ClientOption {
	return func(o *options) {
		o.insecure = insecure
	}
}

// Функция-опция WithTimeout -
func WithTimeout(duration time.Duration) ClientOption {
	return func(o *options) {
		o.timeout = duration
	}
}

// NewClient создает новый клиент HTTP с опциями, которые передаются в качестве аргументов
func NewClient(baseURLStr, username, apiKey string, debug bool, opts ...ClientOption) (*HTTPClient, error) {
	// Парсим, но игнорируем ошибки — будем строить заново
	parsed, err := url.Parse(strings.TrimSpace(baseURLStr))
	if err != nil || parsed.Host == "" {
		return nil, fmt.Errorf("неверный или пустой base URL: %s", baseURLStr)
	}

	// Создаём новый URL только с scheme и host
	cleanURL := &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host, // автоматически обрабатывает порт
	}

	if debug {
		utils.DebugPrint("{client} - Оригинальный baseURL: %s", baseURLStr)
		utils.DebugPrint("{client} - Нормализованный baseURL: %s", cleanURL.String())
	}
	// Создаем конфигурацию с опциями по умолчанию
	cfg := defaultOptions
	for _, o := range opts {
		o(&cfg)
	}
	// Создаем транспорт с нужными опциями
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.insecure,
		},
		TLSHandshakeTimeout: cfg.tlsHandshakeTimeout,
	}
	// Создаем транспорт с Basic Auth, который будет добавляться в каждый запрос
	auth := authTransport{
		username: username,
		apiKey:   apiKey,
		base:     transport,
	}

	return &HTTPClient{
		client: &http.Client{
			Transport: auth,
			Timeout:   cfg.timeout,
		},
		baseURL: cleanURL,
	}, nil
}

// DoRequest — универсальный метод для любого HTTP-запроса
// DoRequest — универсальный метод, формирует URL вручную для TestRail
func (c *HTTPClient) DoRequest(method, endpoint string, body io.Reader, queryParams map[string]string) (*http.Response, error) {
	// Очищаем endpoint от ведущего слеша
	cleanEndpoint := strings.TrimPrefix(endpoint, "/")
	utils.DebugPrint("{DoRequest} - cleanEndpoint: %s", cleanEndpoint)

	// Формируем путь вручную — TestRail требует ? в пути некодированным
	path := apiPrefix + cleanEndpoint
	utils.DebugPrint("{DoRequest} - Path: %s", path)
	// Базовый URL как строка (с trailing слешем, если нужно)
	base := strings.TrimSuffix(c.baseURL.String(), "/")
	utils.DebugPrint("{DoRequest} - Базовый URL: %s", base)
	// Полный URL как строка
	fullURL := base + "/" + path

	// Добавляем query-параметры через & (особенность TestRail)
	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			q.Add(k, v)
		}
		fullURL += "&" + q.Encode() // & вместо ?
	}

	utils.DebugPrint("{DoRequest} - Формируемый URL: %s", fullURL)
	// Создаем сам запрос
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	// Устанавливаем заголовок Content-Type
	req.Header.Set("Content-Type", "application/json")
	// Выпоняем сформированный запрос
	return c.client.Do(req)
}

// Get — обёртка для GET-запросов с умной обработкой ошибок
func (c *HTTPClient) Get(endpoint string, queryParams map[string]string) (*http.Response, error) {
	resp, err := c.DoRequest("GET", endpoint, nil, queryParams)
	if err != nil {
		return nil, err
	}

	// Если статус не OK — сразу форматируем красивую ошибку
	if resp.StatusCode != http.StatusOK {
		return nil, c.formatAPIError(resp)
	}

	return resp, nil
}

// Post — обёртка для POST-запросов с умной обработкой ошибок
func (c *HTTPClient) Post(endpoint string, body io.Reader, queryParams map[string]string) (*http.Response, error) {
	resp, err := c.DoRequest("POST", endpoint, body, queryParams)
	if err != nil {
		return nil, err
	}

	// Аналогично: если не OK — красивая ошибка
	if resp.StatusCode != http.StatusOK {
		return nil, c.formatAPIError(resp)
	}

	return resp, nil
}

// formatAPIError — центральная функция для красивого форматирования ошибок от TestRail
func (c *HTTPClient) formatAPIError(resp *http.Response) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API вернул %s, но не удалось прочитать тело ошибки: %w", resp.Status, err)
	}

	// Пытаемся распарсить как JSON с полем "error"
	var errStruct struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(bodyBytes, &errStruct) == nil && errStruct.Error != "" {
		// Go автоматически декодирует \uXXXX в нормальный UTF-8 текст
		return fmt.Errorf("API вернул %s: %s", resp.Status, errStruct.Error)
	}

	// Если не получилось распарсить как JSON с error — выводим тело как есть
	return fmt.Errorf("API вернул %s: %s", resp.Status, string(bodyBytes))
}
