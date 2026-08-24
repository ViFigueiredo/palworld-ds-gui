package main

import (
	"encoding/json"
	"palworld-ds-gui-server/utils"

	"github.com/gorilla/websocket"
)

// Eventos WebSocket da REST API do Palworld (consistente com API.md).
// Cada evento espelha um endpoint de /v1/api (GET para leitura, POST para ação).
var (
	palworldInfoEvent     = "PALWORLD_INFO"
	palworldPlayersEvent  = "PALWORLD_PLAYERS"
	palworldSettingsEvent = "PALWORLD_SETTINGS"
	palworldMetricsEvent  = "PALWORLD_METRICS"
	palworldGameDataEvent = "PALWORLD_GAMEDATA"
	palworldSaveEvent     = "PALWORLD_SAVE"
	palworldAnnounceEvent = "PALWORLD_ANNOUNCE"
	palworldKickEvent     = "PALWORLD_KICK"
	palworldBanEvent      = "PALWORLD_BAN"
	palworldUnbanEvent    = "PALWORLD_UNBAN"
	palworldShutdownEvent = "PALWORLD_SHUTDOWN"
	palworldStopEvent     = "PALWORLD_STOP"
)

// PalworldAPIResponse — resposta padrão: dados da API do Palworld ou erro.
type PalworldAPIResponse struct {
	BaseResponse
	Data interface{} `json:"data"`
}

func palworldRespond(conn *websocket.Conn, event, eventId string, data interface{}, err error) {
	resp := PalworldAPIResponse{
		BaseResponse: BaseResponse{
			Event:   event,
			EventId: eventId,
			Success: err == nil,
		},
		Data: data,
	}
	if err != nil {
		resp.Error = err.Error()
		utils.Log(err.Error())
	}
	conn.WriteJSON(resp)
}

func palworldEventID(data []byte) string {
	var req BaseRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return ""
	}
	return req.EventId
}

// ---------- LEITURA (GET) ----------

func PalworldInfoHandler(conn *websocket.Conn, data []byte) {
	info, err := PalworldGetInfo()
	palworldRespond(conn, palworldInfoEvent, palworldEventID(data), info, err)
}

func PalworldPlayersHandler(conn *websocket.Conn, data []byte) {
	players, err := PalworldGetPlayers()
	palworldRespond(conn, palworldPlayersEvent, palworldEventID(data), players, err)
}

func PalworldSettingsHandler(conn *websocket.Conn, data []byte) {
	settings, err := PalworldGetSettings()
	palworldRespond(conn, palworldSettingsEvent, palworldEventID(data), settings, err)
}

func PalworldMetricsHandler(conn *websocket.Conn, data []byte) {
	metrics, err := PalworldGetMetrics()
	palworldRespond(conn, palworldMetricsEvent, palworldEventID(data), metrics, err)
}

func PalworldGameDataHandler(conn *websocket.Conn, data []byte) {
	gd, err := PalworldGetGameData()
	palworldRespond(conn, palworldGameDataEvent, palworldEventID(data), gd, err)
}

// ---------- AÇÃO (POST) ----------

type PalworldAnnounceRequest struct {
	BaseRequest
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

type PalworldUserRequest struct {
	BaseRequest
	Data struct {
		UserID  string `json:"userid"`
		Message string `json:"message"`
	} `json:"data"`
}

type PalworldUnbanRequest struct {
	BaseRequest
	Data struct {
		UserID string `json:"userid"`
	} `json:"data"`
}

type PalworldShutdownRequest struct {
	BaseRequest
	Data struct {
		WaitTime int    `json:"waittime"`
		Message  string `json:"message"`
	} `json:"data"`
}

func PalworldSaveHandler(conn *websocket.Conn, data []byte) {
	err := PalworldSave()
	palworldRespond(conn, palworldSaveEvent, palworldEventID(data), "world saved", err)
}

func PalworldAnnounceHandler(conn *websocket.Conn, data []byte) {
	var req PalworldAnnounceRequest
	if err := json.Unmarshal(data, &req); err != nil {
		palworldRespond(conn, palworldAnnounceEvent, palworldEventID(data), nil, err)
		return
	}
	err := PalworldAnnounce(req.Data.Message)
	palworldRespond(conn, palworldAnnounceEvent, req.EventId, "announced", err)
}

func PalworldKickHandler(conn *websocket.Conn, data []byte) {
	var req PalworldUserRequest
	if err := json.Unmarshal(data, &req); err != nil {
		palworldRespond(conn, palworldKickEvent, palworldEventID(data), nil, err)
		return
	}
	err := PalworldKick(req.Data.UserID, req.Data.Message)
	palworldRespond(conn, palworldKickEvent, req.EventId, "player kicked", err)
}

func PalworldBanHandler(conn *websocket.Conn, data []byte) {
	var req PalworldUserRequest
	if err := json.Unmarshal(data, &req); err != nil {
		palworldRespond(conn, palworldBanEvent, palworldEventID(data), nil, err)
		return
	}
	err := PalworldBan(req.Data.UserID, req.Data.Message)
	palworldRespond(conn, palworldBanEvent, req.EventId, "player banned", err)
}

func PalworldUnbanHandler(conn *websocket.Conn, data []byte) {
	var req PalworldUnbanRequest
	if err := json.Unmarshal(data, &req); err != nil {
		palworldRespond(conn, palworldUnbanEvent, palworldEventID(data), nil, err)
		return
	}
	err := PalworldUnban(req.Data.UserID)
	palworldRespond(conn, palworldUnbanEvent, req.EventId, "player unbanned", err)
}

func PalworldShutdownHandler(conn *websocket.Conn, data []byte) {
	var req PalworldShutdownRequest
	if err := json.Unmarshal(data, &req); err != nil {
		palworldRespond(conn, palworldShutdownEvent, palworldEventID(data), nil, err)
		return
	}
	err := PalworldShutdown(req.Data.WaitTime, req.Data.Message)
	palworldRespond(conn, palworldShutdownEvent, req.EventId, "server shutdown scheduled", err)
}

func PalworldStopHandler(conn *websocket.Conn, data []byte) {
	err := PalworldStop()
	palworldRespond(conn, palworldStopEvent, palworldEventID(data), "server stopped", err)
}
