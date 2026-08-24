package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// PalworldAPIConfig — configuração da REST API do servidor Palworld.
// Consistente com a documentação em API.md (§2 Autenticação, §3 Base URL).
// Valores podem ser sobrescritos por variáveis de ambiente:
//
//	PALWORLD_API_URL      (default: http://localhost:8212/v1/api)
//	PALWORLD_API_USERNAME (default: admin)
//	PALWORLD_API_PASSWORD (obrigatória — mesma ADMIN_PASSWORD do servidor)
type PalworldAPIConfig struct {
	BaseURL  string
	Username string
	Password string
}

var palworldAPI = PalworldAPIConfig{
	BaseURL:  "http://localhost:8212/v1/api",
	Username: "admin",
	Password: os.Getenv("PALWORLD_API_PASSWORD"),
}

func init() {
	if v := os.Getenv("PALWORLD_API_URL"); v != "" {
		palworldAPI.BaseURL = v
	}
	if v := os.Getenv("PALWORLD_API_USERNAME"); v != "" {
		palworldAPI.Username = v
	}
}

// ---------- STRUCTS (respostas conforme API.md) ----------

type PalworldInfo struct {
	Version     string `json:"version"`
	ServerName  string `json:"servername"`
	Description string `json:"description"`
	WorldGUID   string `json:"worldguid"`
}

type PalworldPlayer struct {
	PlayerID  string  `json:"playerid"`
	SteamID   string  `json:"steamid"`
	Name      string  `json:"name"`
	Level     int     `json:"level"`
	Ping      int     `json:"ping"`
	LocationX float64 `json:"location_x"`
	LocationY float64 `json:"location_y"`
	LocationZ float64 `json:"location_z"`
}

type PalworldPlayersResponse struct {
	Players []PalworldPlayer `json:"players"`
}

type PalworldMetrics struct {
	CurrentPlayerNum int     `json:"currentplayernum"`
	ServerFPS        int     `json:"serverfps"`
	ServerFPSAverage float64 `json:"serverfpsaverage"`
	ServerFrameTime  float64 `json:"serverframetime"`
	Days             int     `json:"days"`
	MaxPlayerNum     int     `json:"maxplayernum"`
	BaseCampNum      int     `json:"basecampnum"`
	Uptime           int     `json:"uptime"`
}

// PalworldSettings é o mapa completo de configurações do mundo (schema em API.md §5.3).
type PalworldSettings map[string]interface{}

// PalworldGameData é o snapshot de atores do mundo.
// Requer ENABLE_GAMEDATA_API=true no servidor; caso contrário a API responde 404 (ver API.md §5.5).
type PalworldGameData map[string]interface{}

// ---------- HTTP ----------

var palworldHTTPClient = &http.Client{Timeout: 15 * time.Second}

func palworldRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, palworldAPI.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(palworldAPI.Username, palworldAPI.Password)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := palworldHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("palworld api: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return data, nil
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("palworld api: unauthorized (401) — verifique PALWORLD_API_USERNAME/PALWORLD_API_PASSWORD")
	case http.StatusBadRequest:
		return nil, fmt.Errorf("palworld api: bad request (400): %s", strings.TrimSpace(string(data)))
	default:
		return nil, fmt.Errorf("palworld api: error (%d): %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
}

// ---------- GET ----------

func PalworldGetInfo() (*PalworldInfo, error) {
	data, err := palworldRequest(http.MethodGet, "/info", nil)
	if err != nil {
		return nil, err
	}
	var info PalworldInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func PalworldGetPlayers() (*PalworldPlayersResponse, error) {
	data, err := palworldRequest(http.MethodGet, "/players", nil)
	if err != nil {
		return nil, err
	}
	var players PalworldPlayersResponse
	if err := json.Unmarshal(data, &players); err != nil {
		return nil, err
	}
	return &players, nil
}

func PalworldGetSettings() (*PalworldSettings, error) {
	data, err := palworldRequest(http.MethodGet, "/settings", nil)
	if err != nil {
		return nil, err
	}
	var settings PalworldSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func PalworldGetMetrics() (*PalworldMetrics, error) {
	data, err := palworldRequest(http.MethodGet, "/metrics", nil)
	if err != nil {
		return nil, err
	}
	var metrics PalworldMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil
}

func PalworldGetGameData() (*PalworldGameData, error) {
	data, err := palworldRequest(http.MethodGet, "/game-data", nil)
	if err != nil {
		return nil, err
	}
	var gd PalworldGameData
	if err := json.Unmarshal(data, &gd); err != nil {
		return nil, err
	}
	return &gd, nil
}

// ---------- POST ----------

func PalworldAnnounce(message string) error {
	_, err := palworldRequest(http.MethodPost, "/announce", map[string]string{"message": message})
	return err
}

func PalworldKick(userid, message string) error {
	body := map[string]string{"userid": userid}
	if message != "" {
		body["message"] = message
	}
	_, err := palworldRequest(http.MethodPost, "/kick", body)
	return err
}

func PalworldBan(userid, message string) error {
	body := map[string]string{"userid": userid}
	if message != "" {
		body["message"] = message
	}
	_, err := palworldRequest(http.MethodPost, "/ban", body)
	return err
}

func PalworldUnban(userid string) error {
	_, err := palworldRequest(http.MethodPost, "/unban", map[string]string{"userid": userid})
	return err
}

func PalworldSave() error {
	_, err := palworldRequest(http.MethodPost, "/save", map[string]string{})
	return err
}

func PalworldShutdown(waittime int, message string) error {
	body := map[string]interface{}{"waittime": waittime}
	if message != "" {
		body["message"] = message
	}
	_, err := palworldRequest(http.MethodPost, "/shutdown", body)
	return err
}

func PalworldStop() error {
	_, err := palworldRequest(http.MethodPost, "/stop", map[string]string{})
	return err
}
