# Palworld REST API — Documentação Completa

Documentação da API REST do servidor dedicado **Palworld** (imagem `thijsvanloef/palworld-server-docker`), verificada no servidor **"Figcodes"** em 24/08/2026 (versão do jogo `v1.0.3.101283`).

Fonte oficial: <https://docs.palworldgame.com/category/rest-api>

---

## 1. Visão geral

A API REST permite ler o estado do servidor e executar ações de administração via HTTP/JSON:

- **Leitura**: informações do servidor, jogadores, configurações do mundo, métricas (FPS, uptime), snapshot do mundo.
- **Ação**: salvar o mundo, mandar mensagens (broadcast), expulsar/banir jogadores, desligar o servidor.

> ⚠️ **Aviso oficial da Pocketpair**: esta API **não foi projetada para ser exposta diretamente à internet** — exposição pública permite manipulação não autorizada do servidor. No nosso ambiente ela está exposta **somente para os 2 IPs autorizados** (mesmos do SSH), com firewall bloqueando o resto do mundo.

### Estado no nosso servidor
| Item | Valor |
|---|---|
| Versão do jogo | `v1.0.3.101283` |
| Nome do servidor | `Figcodes` |
| API REST | Habilitada (`REST_API_ENABLED=true`) |
| Porta interna | `8212` (TCP) |
| Acesso externo | Restrito a `179.154.239.137` e `187.115.158.250` |

---

## 2. Autenticação

**HTTP Basic Auth**

| Campo | Valor |
|---|---|
| Username | `admin` |
| Password | `ADMIN_PASSWORD` do stack (atualmente `M2ZnjtL0F?ixafxS=#`) |

- Sem credenciais ou credenciais erradas → **HTTP 401**.
- A senha é a mesma `ADMIN_PASSWORD` usada para o console de administração.

---

## 3. Base URL

| Origem | URL |
|---|---|
| De dentro do servidor (host) | `http://localhost:8212/v1/api` |
| De fora (somente IPs autorizados) | `http://194.34.232.101:8212/v1/api` |

Exemplo de chamada mínima:

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/info
```

---

## 4. Resumo dos endpoints

| Método | Endpoint | Descrição | Body |
|---|---|---|---|
| GET | `/v1/api/info` | Informações do servidor | — |
| GET | `/v1/api/players` | Lista de jogadores online | — |
| GET | `/v1/api/settings` | Configurações do mundo | — |
| GET | `/v1/api/metrics` | Métricas (FPS, players, uptime, dias) | — |
| GET | `/v1/api/game-data` | Snapshot de todos os atores do mundo | — (requer flag) |
| POST | `/v1/api/announce` | Envia mensagem para todos os jogadores | `{"message": string}` |
| POST | `/v1/api/kick` | Expulsa um jogador | `{"userid": string, "message"?: string}` |
| POST | `/v1/api/ban` | Bane um jogador | `{"userid": string, "message"?: string}` |
| POST | `/v1/api/unban` | Remove o banimento | `{"userid": string}` |
| POST | `/v1/api/save` | Salva o mundo manualmente | — |
| POST | `/v1/api/shutdown` | Desliga o servidor com aviso | `{"waittime": int, "message"?: string}` |
| POST | `/v1/api/stop` | Força a parada imediata do servidor | — |

---

## 5. Endpoints de leitura

### 5.1 `GET /v1/api/info` — Informações do servidor

**Resposta real (verificada):**

```json
{
    "version": "v1.0.3.101283",
    "servername": "Figcodes",
    "description": "Servidor Palworld na Figcodes Soluções",
    "worldguid": "D81C0A8E9EC64000ABDCB6F3C372FEBC"
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `version` | string | Versão do servidor/jogo |
| `servername` | string | Nome do servidor |
| `description` | string | Descrição exibida na lista |
| `worldguid` | string | GUID do mundo salvo (identifica a "seed" do save) |

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/info
```

---

### 5.2 `GET /v1/api/players` — Lista de jogadores

**Resposta real (servidor sem ninguém online):**

```json
{
    "players": []
}
```

**Schema** (`players` é um array de objetos; campos típicos de cada jogador — preenchidos quando há alguém online):

| Campo | Tipo | Descrição |
|---|---|---|
| `playerid` | string | UUID do jogador no servidor (formato `xxxxxxxx-xxxx-...`) |
| `steamid` | string | SteamID numérica do jogador |
| `name` | string | Nome do jogador |
| `level` | int | Nível |
| `ping` | int | Ping do jogador |
| `location_x` / `location_y` / `location_z` | float | Posição no mundo |

> O campo **`playerid`** (UUID) é o **`userid`** exigido pelos endpoints de kick/ban/unban — ver seção 8.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/players
```

---

### 5.3 `GET /v1/api/settings` — Configurações do mundo

Retorna **todas** as configurações ativas do servidor (as mesmas do arquivo `PalWorldSettings.ini`). Campos principais (schema oficial):

| Campo | Tipo |
|---|---|
| `Difficulty` | string (`None`/`Casual`/`Normal`/`Hard`) |
| `DayTimeSpeedRate` | number |
| `NightTimeSpeedRate` | number |
| `ExpRate` | number |
| `PalCaptureRate` | number |
| `PalSpawnNumRate` | number |
| `PalDamageRateAttack` / `PalDamageRateDefense` | number |
| `PlayerDamageRateAttack` / `PlayerDamageRateDefense` | number |
| `PlayerStomachDecreaceRate` / `PlayerStaminaDecreaceRate` | number |
| `PlayerAutoHPRegeneRate` | number |
| `PalStomachDecreaceRate` / `PalStaminaDecreaceRate` | number |
| `PalAutoHPRegeneRate` | number |
| `BuildObjectDamageRate` / `BuildObjectDeteriorationDamageRate` | number |
| `CollectionDropRate` / `CollectionObjectHpRate` / `CollectionObjectRespawnSpeedRate` | number |
| `EnemyDropItemRate` | number |
| `DeathPenalty` | string |
| `bEnablePlayerToPlayerDamage` / `bEnableFriendlyFire` / `bEnableInvaderEnemy` | boolean |
| `bEnableAimAssistPad` / `bEnableAimAssistKeyboard` | boolean |
| `DropItemMaxNum` / `DropItemMaxNum_UNKO` | number |
| `BaseCampMaxNum` / `BaseCampWorkerMaxNum` | number |
| `DropItemAliveMaxHours` | number |
| `bAutoResetGuildNoOnlinePlayers` | boolean |
| `AutoResetGuildTimeNoOnlinePlayers` | number |
| `GuildPlayerMaxNum` | number |
| `PalEggDefaultHatchingTime` | number |
| `WorkSpeedRate` | number |
| `bIsMultiplay` / `bIsPvP` | boolean |
| `bCanPickupOtherGuildDeathPenaltyDrop` / `bEnableNonLoginPenalty` / `bEnableFastTravel` | boolean |
| `bIsStartLocationSelectByMap` / `bExistPlayerAfterLogout` / `bEnableDefenseOtherGuildPlayer` | boolean |
| `CoopPlayerMaxNum` / `ServerPlayerMaxNum` | number |
| `ServerName` / `ServerDescription` | string |
| `PublicPort` / `PublicIP` | number / string |
| `RCONEnabled` / `RCONPort` | boolean / number |
| `Region` | string |
| `bUseAuth` | boolean |
| `BanListURL` | string |
| `RESTAPIEnabled` / `RESTAPIPort` | boolean / number |
| `bShowPlayerList` | boolean |
| `AllowConnectPlatform` | string |
| `bIsUseBackupSaveData` | boolean |
| `LogFormatType` | string |

**Uso prático**: permite à GUI **exibir e auditar** a configuração do mundo (dificuldade, taxas de EXP/captura, PvP on/off, limites de jogadores/guildas) sem abrir arquivos.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/settings
```

---

### 5.4 `GET /v1/api/metrics` — Métricas do servidor

**Resposta real (verificada):**

```json
{
    "currentplayernum": 0,
    "serverfps": 59,
    "serverfpsaverage": 59.400001525878906,
    "serverframetime": 16.82914924621582,
    "days": 6,
    "maxplayernum": 10,
    "basecampnum": 1,
    "uptime": 628
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `currentplayernum` | int | Jogadores online agora |
| `serverfps` | int | FPS do servidor (tick rate) |
| `serverfpsaverage` | float | FPS médio |
| `serverframetime` | float | Tempo de quadro (ms) |
| `days` | int | Dias dentro do jogo |
| `maxplayernum` | int | Máximo de jogadores configurado |
| `basecampnum` | int | Número de acampamentos no mundo |
| `uptime` | int | Uptime do servidor em segundos |

> `serverfps` em ~59/60 indica servidor saudável. Se cair muito (ex.: < 30), o servidor está com lag de processamento — os limites de recursos do stack (4–6 vCPU, 16 GB) existem justamente para evitar isso.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/metrics
```

---

### 5.5 `GET /v1/api/game-data` — Snapshot de atores do mundo

Retorna um snapshot com **todos os atores** do mundo (personagens, pals, caixas de pals) no momento da chamada, com posição, HP, nível, guilda etc.

**Estado atual no nosso servidor**: retorna **HTTP 404** com a mensagem `PalGameDataBridge GameData API is not enabled` — porque a flag `-enable-gamedata-api` está desligada por padrão.

**Para habilitar**: adicionar ao stack (na imagem `palworld-server-docker` ≥ v2.1.0):

```yaml
      ENABLE_GAMEDATA_API: "true"
```

e dar "Update the stack" no Portainer.

**Schema da resposta (quando habilitado)**:

| Campo | Tipo | Descrição |
|---|---|---|
| `Time` | string | Timestamp ("YYYY-MM-DD HH:MM:SS", hora local do servidor) |
| `FPS` | float | FPS instantâneo |
| `AverageFPS` | float | FPS médio |
| `ActorData` | array | Atores (`Character` / `PalBox`) |

Campos de cada ator do tipo `Character`: `Type`, `InstanceID`, `UnitType` (`Player`/`OtomoPal`/`BaseCampPal`/`WildPal`/`NPC`), `NickName`, `TrainerInstanceID`, `TrainerNickName`, `userid` (só player), `ip`, `level`, `HP`, `MaxHP`, `GuildID`, `GuildName`, `Class`, `Action`, `AI_Action`, `LocationX/Y/Z`, `RotationX/Y/Z`, `Stage`, `IsActive`.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" http://194.34.232.101:8212/v1/api/game-data
```

> Aviso: com servidor grande, essa resposta pode ser **pesada** (centenas de atores). Não chamar em loop.

---

## 6. Endpoints de ação

### 6.1 `POST /v1/api/announce` — Mensagem para todos (broadcast)

**Body** (obrigatório):

```json
{ "message": "Servidor reinicia em 10 minutos!" }
```

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `message` | string | ✅ | Texto exibido para todos os jogadores |

**Verificado**: sem `message` → HTTP 400 `Bad Request`; com `message` → HTTP 200 (mensagem exibida no jogo).

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/announce \
  -H "Content-Type: application/json" \
  -d '{"message":"Servidor reinicia em 10 minutos!"}'
```

---

### 6.2 `POST /v1/api/kick` — Expulsar jogador

**Body** (obrigatório):

```json
{ "userid": "3a6f2c11-9b44-4f3e-8a02-1d5b7c9e4f21", "message": "Motivo: regras" }
```

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `userid` | string | ✅ | UUID do jogador (campo `playerid` de `/players`) |
| `message` | string | ❌ | Mensagem exibida ao jogador expulso |

**Verificado**: `userid` inexistente → HTTP 400 `Bad Request`. Sucesso → HTTP 200.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/kick \
  -H "Content-Type: application/json" \
  -d '{"userid":"<uuid-do-jogador>","message":"Expulso"}'
```

---

### 6.3 `POST /v1/api/ban` — Banir jogador

**Body** (obrigatório):

```json
{ "userid": "3a6f2c11-9b44-4f3e-8a02-1d5b7c9e4f21", "message": "Banido por comportamento" }
```

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `userid` | string | ✅ | UUID do jogador (campo `playerid` de `/players`) |
| `message` | string | ❌ | Mensagem exibida ao jogador banido |

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/ban \
  -H "Content-Type: application/json" \
  -d '{"userid":"<uuid-do-jogador>","message":"Banido"}'
```

---

### 6.4 `POST /v1/api/unban` — Remover banimento

**Body** (obrigatório):

```json
{ "userid": "3a6f2c11-9b44-4f3e-8a02-1d5b7c9e4f21" }
```

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `userid` | string | ✅ | UUID do jogador |

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/unban \
  -H "Content-Type: application/json" \
  -d '{"userid":"<uuid-do-jogador>"}'
```

---

### 6.5 `POST /v1/api/save` — Salvar o mundo

Sem body. **Verificado**: HTTP 200 (salva o mundo imediatamente). Uso recomendado antes de desligar/reiniciar.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/save
```

---

### 6.6 `POST /v1/api/shutdown` — Desligar com aviso

**Body** (obrigatório):

```json
{ "waittime": 60, "message": "Servidor desligando para manutenção" }
```

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `waittime` | int | ✅ | Segundos de espera antes de desligar |
| `message` | string | ❌ | Mensagem exibida aos jogadores |

> O container sobe de novo automaticamente (política de restart do swarm) — no caso de manutenção, derrubar o serviço no Portainer depois.

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/shutdown \
  -H "Content-Type: application/json" \
  -d '{"waittime":60,"message":"Desligando em 1 minuto"}'
```

---

### 6.7 `POST /v1/api/stop` — Forçar parada imediata

Sem body. **Para uso de emergência** — encerra o processo do servidor na hora (risco de perda de progresso; prefira `save` + `shutdown`).

```bash
curl -u "admin:M2ZnjtL0F?ixafxS=#" -X POST http://194.34.232.101:8212/v1/api/stop
```

---

## 7. Respostas de erro

| Código | Significado | Situações típicas |
|---|---|---|
| `200` | Sucesso | Ação executada / dados retornados |
| `400` | Bad Request | Body ausente, campo obrigatório faltando (`announce` sem `message`, `kick`/`unban` com `userid` inválido) |
| `401` | Unauthorized | Credenciais ausentes ou erradas |
| `404` | Not Found | Endpoint não habilitado (ex.: `game-data` sem `ENABLE_GAMEDATA_API`) |

---

## 8. ⚠️ Pegadinha: `userid` ≠ SteamID

- Os endpoints `kick`, `ban` e `unban` esperam o campo **`userid`** — o **UUID do jogador no servidor** (formato `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`), **não** o SteamID numérico.
- O `userid` é obtido pelo campo **`playerid`** da resposta de `GET /v1/api/players` (ou pelo campo `userid` de cada ator `Player` em `game-data`).
- (O README da imagem cita "kick {SteamID}", mas a documentação oficial da Pocketpair e o comportamento real usam `userid` — confirmado: enviar UUID inválido retorna 400, não kicka ninguém.)
- **Fluxo correto na GUI**: `GET /players` → guardar `playerid` → usar como `userid` no kick/ban/unban.

---

## 9. Uso interno (dentro do servidor)

A imagem oferece o comando `rest-cli` dentro do container (não precisa expor porta):

```bash
# pegar o container do serviço
CID=$(docker ps --filter name=palworld_palworld -q | head -1)

docker exec -i "$CID" rest-cli info
docker exec -i "$CID" rest-cli players
docker exec -i "$CID" rest-cli save
docker exec -i "$CID" rest-cli announce '{"message":"Ola!"}'
docker exec -i "$CID" rest-cli kick '{"userid":"<uuid>"}'
docker exec -i "$CID" rest-cli ban '{"userid":"<uuid>","message":"..."}'
docker exec -i "$CID" rest-cli unban '{"userid":"<uuid>"}'
docker exec -i "$CID" rest-cli shutdown '{"waittime":60,"message":"..."}'
docker exec -i "$CID" rest-cli stop
```

---

## 10. O que podemos fazer com a API (casos de uso para a GUI)

| Recurso da GUI | Endpoints usados |
|---|---|
| **Painel de status** (tela inicial) | `GET /info` + `GET /metrics` (versão, jogadores online, FPS, uptime, dias) |
| **Lista de jogadores online** (nome, nível, ping, posição) | `GET /players` |
| **Expulsar / banir / desbanir** com motivo | `POST /kick` / `POST /ban` / `POST /unban` |
| **Broadcast** (avisos, recados, anúncios) | `POST /announce` |
| **Salvar mundo** (botão) | `POST /save` |
| **Reinício com aviso** (manutenção) | `POST /shutdown` com `waittime` + mensagem |
| **Parada de emergência** | `POST /stop` |
| **Monitoramento de saúde** (alerta se FPS cair) | `GET /metrics` (polling) |
| **Exibir configuração do mundo** (dificuldade, taxas, PvP) | `GET /settings` |
| **Mapa / inventário do mundo** (avançado) | `GET /game-data` (após `ENABLE_GAMEDATA_API=true`) |

**Observações de integração para a GUI:**
- Autenticação: `Authorization: Basic base64(admin:SENHA)` — a senha deve ficar em variável de ambiente/config, não no código.
- Polling recomendado: `metrics`/`players` a cada 5–10 s; `game-data` **nunca** em loop (resposta pesada).
- Sempre obter `userid` da resposta de `/players` (nunca pedir SteamID ao usuário para kick/ban).

---

## 11. Estado atual do servidor (verificado em 24/08/2026)

| Métrica | Valor |
|---|---|
| Versão | `v1.0.3.101283` |
| Nome | `Figcodes` |
| Mundo | `D81C0A8E9EC64000ABDCB6F3C372FEBC` |
| FPS do servidor | ~59 (saudável) |
| Máx. jogadores | 10 |
| Uptime na verificação | ~10 min |
| Acesso externo à API | Somente IPs `179.154.239.137` e `187.115.158.250` |

---

## 12. Segurança

1. **Não exponha a API publicamente** (aviso oficial). No nosso caso, o firewall (mangle PREROUTING) descarta qualquer origem fora dos 2 IPs autorizados — mesmo com a porta 8212 publicada no swarm.
2. A `ADMIN_PASSWORD` é a chave-mestra (API + console): mantenha forte e fora do código.
3. Prefira `shutdown` com `waittime` a `stop` (evita corromper o save).
4. Use `save` antes de qualquer operação de manutenção.
5. Se a GUI rodar **dentro da mesma VPS** (ex.: container na mesma rede), use `localhost:8212` e mantenha o firewall — dupla proteção.

---

## 13. Referências

- Documentação oficial da API: <https://docs.palworldgame.com/category/rest-api>
- Repositório da imagem: <https://github.com/thijsvanloef/palworld-server-docker>
- Guia do servidor Palworld: <https://docs.palworldgame.com>
