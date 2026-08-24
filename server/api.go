package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"palworld-ds-gui-server/utils"
	"path"
	"path/filepath"
)

type Api struct {
}

func NewApi() *Api {
	return &Api{}
}

func PrintApiKey() {
	fmt.Printf("\n\n🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨\n")
	fmt.Printf("Anyone with your API key can control your server. Keep it secret!\n")
	fmt.Printf("Your current API key is: %s\n", utils.Settings.General.APIKey)
	fmt.Printf("To generate a new API key, run the program with the --newkey flag\n")
	fmt.Printf("🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨\n\n")
}

func (a *Api) Init() {
	if !HasApiKey() || utils.Launch.ForceNewKey {
		GenerateApiKey()
		utils.SaveSettings()
	}

	if utils.Launch.ShowKey {
		PrintApiKey()
	}

	internalIp := utils.GetOutboundIP()
	externalIp, err := utils.GetExternalIPv4()
	if err != nil {
		utils.LogToFile("Failed to get external IP: "+err.Error(), true)
		fmt.Printf("Server is running on %s:%d\n", internalIp, utils.Launch.Port)
	} else {
		fmt.Printf("Server is running on %s:%d (Local IP: %s:%d)\n", externalIp, utils.Launch.Port, internalIp, utils.Launch.Port)
	}

	utils.EmitConsoleLog = LogToClient

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/backups/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		apiKey := r.Header.Get("Authorization")

		if apiKey != utils.Settings.General.APIKey {
			utils.Log(fmt.Sprintf("Unauthorized backup download from %s", r.RemoteAddr))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract the specific file path from the URL.
		filePath := r.URL.Path[len("/backups/"):]

		// Safely join the base path with the requested file path to avoid directory traversal attacks.
		safeFilePath := path.Join(utils.Config.BackupsPath, filePath)

		// Ensure the file exists and is not a directory before serving.
		if info, err := os.Stat(safeFilePath); err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		// Serve the requested file.
		http.ServeFile(w, r, safeFilePath)
	})

	// Web UI (SPA): serve o frontend buildado (ver Dockerfile) na raiz.
	// O caminho pode ser sobrescrito por PALWORLD_GUI_WEB_DIR.
	webDir := os.Getenv("PALWORLD_GUI_WEB_DIR")
	if webDir == "" {
		webDir = "/usr/local/share/palworld-ds-gui-web"
	}
	http.Handle("/", spaFileServer(webDir))

	http.ListenAndServe(fmt.Sprintf(":%d", utils.Launch.Port), nil)
}

// spaFileServer serve os arquivos estáticos da SPA com fallback para index.html
// (routing do lado do cliente — React Router).
func spaFileServer(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(p); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func GenerateApiKey() {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hasher := sha256.New()
	hasher.Write(randomBytes)
	hash := hasher.Sum(nil)
	stringHash := hex.EncodeToString(hash)

	utils.Settings.General.APIKey = stringHash
	err := utils.SaveSettings()
	if err != nil {
		panic(err)
	}

	PrintApiKey()
}

func HasApiKey() bool {
	if utils.Settings.General.APIKey == "" || utils.Settings.General.APIKey == "CHANGE_ME" {
		return false
	}

	return true
}
