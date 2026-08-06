# Uzinha - Painel de Controle para Minions

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Criar um painel web simples ("Uzinha") para monitorar e gerenciar múltiplos minions remotamente, visualizando dados coletados via API.

**Architecture:** Aplicação web standalone em Go com HTML estático. Serve um dashboard que conecta em múltiplos minions via API, coleta dados e exibe em tempo real. Sem banco de dados - configuração via JSON.

**Tech Stack:** Go (net/http), HTML/CSS/JS vanilla, API do minion-agent

## Global Constraints

- Comunicação via HTTPS com TLS self-signed (ignorar verificação)
- Autenticação via API key por minion
- Configuração de minions em arquivo JSON
- Sem dependências externas (Go stdlib + HTML estático)
- Porta padrão: 8080

---

## Estrutura de Arquivos

| Arquivo | Responsabilidade |
|---|---|
| `uzinha/main.go` | Servidor web, rotas, lógica de coleta |
| `uzinha/config.json` | Lista de minions para monitorar |
| `uzinha/static/index.html` | Dashboard principal |
| `uzinha/static/style.css` | Estilos |
| `uzinha/static/app.js` | Lógica do frontend |
| `docs/uzinha.md` | Documentação de uso |

---

### Task 1: Configuração dos Minions

**Files:**
- Create: `uzinha/config.json`

**Interfaces:**
- Consumes: N/A (arquivo de configuração inicial)
- Produz: Lista de minions com IP, porta e API key

- [ ] **Step 1: Criar arquivo de configuração**

```json
{
  "minions": [
    {
      "name": "minion-local",
      "host": "https://127.0.0.1:9870",
      "api_key": "minion_sk_SUA_API_KEY_AQUI",
      "insecure": true
    }
  ],
  "server": {
    "port": 8080
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add uzinha/config.json
git commit -m "feat(uzinha): add minion configuration file"
```

---

### Task 2: Servidor Web Go

**Files:**
- Create: `uzinha/main.go`

**Interfaces:**
- Consumes: `uzinha/config.json`
- Produz: Servidor HTTP na porta 8080 com endpoints para API e arquivos estáticos

- [ ] **Step 1: Criar main.go**

```go
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type MinionConfig struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	APIKey  string `json:"api_key"`
	Insecure bool  `json:"insecure"`
}

type Config struct {
	Minions []MinionConfig `json:"minions"`
	Server  struct {
		Port int `json:"port"`
	} `json:"server"`
}

type MinionData struct {
	Name    string          `json:"name"`
	Host    string          `json:"host"`
	Online  bool            `json:"online"`
	Agent   json.RawMessage `json:"agent,omitempty"`
	System  json.RawMessage `json:"system,omitempty"`
	Memory  json.RawMessage `json:"memory,omitempty"`
	Disk    json.RawMessage `json:"disk,omitempty"`
	Users   json.RawMessage `json:"users,omitempty"`
	Error   string          `json:"error,omitempty"`
}

var config Config

func main() {
	configData, err := os.ReadFile("uzinha/config.json")
	if err != nil {
		log.Fatalf("Erro ao ler config: %v", err)
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Erro ao parsear config: %v", err)
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/minions", handleMinions)
	http.HandleFunc("/api/minion/", handleMinionDetail)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("uzinha/static"))))

	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Uzinha rodando em http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "uzinha/static/index.html")
}

func handleMinions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := make([]MinionData, 0, len(config.Minions))
	for _, m := range config.Minions {
		md := fetchMinionData(m)
		data = append(data, md)
	}

	json.NewEncoder(w).Encode(data)
}

func handleMinionDetail(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	for _, m := range config.Minions {
		if m.Name == name {
			w.Header().Set("Content-Type", "application/json")
			md := fetchMinionFullData(m)
			json.NewEncoder(w).Encode(md)
			return
		}
	}

	http.Error(w, "minion not found", http.StatusNotFound)
}

func fetchMinionData(m MinionConfig) MinionData {
	md := MinionData{Name: m.Name, Host: m.Host}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: m.Insecure},
		},
	}

	req, _ := http.NewRequest("GET", m.Host+"/api/v1/agent", nil)
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		md.Error = err.Error()
		return md
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		md.Error = fmt.Sprintf("status %d", resp.StatusCode)
		return md
	}

	md.Online = true
	body, _ := io.ReadAll(resp.Body)
	md.Agent = body

	return md
}

func fetchMinionFullData(m MinionConfig) MinionData {
	md := fetchMinionData(m)
	if !md.Online {
		return md
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: m.Insecure},
		},
	}

	endpoints := map[string]*json.RawMessage{
		"/api/v1/system": &md.System,
		"/api/v1/memory": &md.Memory,
		"/api/v1/disk":   &md.Disk,
		"/api/v1/users":  &md.Users,
	}

	for ep, target := range endpoints {
		req, _ := http.NewRequest("GET", m.Host+ep, nil)
		req.Header.Set("Authorization", "Bearer "+m.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			*target = body
		}
	}

	return md
}
```

- [ ] **Step 2: Verificar compilação**

```bash
cd uzinha && go mod init uzinha && go build -o uzinha .
```

- [ ] **Step 3: Commit**

```bash
git add uzinha/main.go uzinha/go.mod uzinha/go.sum
git commit -m "feat(uzinha): add web server with minion data collection"
```

---

### Task 3: Dashboard HTML

**Files:**
- Create: `uzinha/static/index.html`

**Interfaces:**
- Consumes: API `/api/minions` e `/api/minion/?name=...`
- Produz: Dashboard responsivo com cards de minions

- [ ] **Step 1: Criar index.html**

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Uzinha - Minion Control</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <h1>🍳 Uzinha</h1>
        <p>Minion Control Panel</p>
    </header>

    <main>
        <section id="minions-grid" class="minions-grid">
            <div class="loading">Carregando minions...</div>
        </section>

        <section id="minion-detail" class="minion-detail hidden">
            <button class="back-btn" onclick="showGrid()">← Voltar</button>
            <div id="detail-content"></div>
        </section>
    </main>

    <script src="/static/app.js"></script>
</body>
</html>
```

- [ ] **Step 2: Commit**

```bash
git add uzinha/static/index.html
git commit -m "feat(uzinha): add dashboard HTML"
```

---

### Task 4: Estilos CSS

**Files:**
- Create: `uzinha/static/style.css`

**Interfaces:**
- Consumes: HTML do index.html
- Produz: Dashboard escuro com cards coloridos

- [ ] **Step 1: Criar style.css**

```css
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #1a1a2e;
    color: #eee;
    min-height: 100vh;
}

header {
    background: #16213e;
    padding: 1rem 2rem;
    border-bottom: 2px solid #0f3460;
}

header h1 {
    font-size: 1.8rem;
    color: #e94560;
}

header p {
    color: #888;
    font-size: 0.9rem;
}

main {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
}

.minions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
}

.loading {
    grid-column: 1 / -1;
    text-align: center;
    color: #888;
    padding: 3rem;
}

.minion-card {
    background: #16213e;
    border-radius: 12px;
    padding: 1.5rem;
    border: 1px solid #0f3460;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
}

.minion-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 25px rgba(233, 69, 96, 0.2);
}

.minion-card.offline {
    opacity: 0.5;
    border-color: #666;
}

.minion-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.minion-name {
    font-size: 1.2rem;
    font-weight: 600;
}

.status-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #4ade80;
}

.status-dot.offline {
    background: #ef4444;
}

.minion-info {
    font-size: 0.85rem;
    color: #aaa;
}

.minion-info p {
    margin: 0.3rem 0;
}

.minion-detail {
    background: #16213e;
    border-radius: 12px;
    padding: 2rem;
    border: 1px solid #0f3460;
}

.minion-detail.hidden {
    display: none;
}

.back-btn {
    background: #0f3460;
    color: #eee;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    margin-bottom: 1.5rem;
}

.back-btn:hover {
    background: #1a4a7a;
}

.detail-section {
    margin-bottom: 1.5rem;
}

.detail-section h3 {
    color: #e94560;
    margin-bottom: 0.8rem;
    font-size: 1rem;
    text-transform: uppercase;
    letter-spacing: 1px;
}

.detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1rem;
}

.detail-item {
    background: #1a1a2e;
    padding: 1rem;
    border-radius: 8px;
}

.detail-item label {
    display: block;
    color: #888;
    font-size: 0.75rem;
    text-transform: uppercase;
    margin-bottom: 0.3rem;
}

.detail-item span {
    font-size: 1.1rem;
    font-weight: 500;
}

.json-view {
    background: #0d1117;
    padding: 1rem;
    border-radius: 8px;
    overflow-x: auto;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 0.8rem;
    color: #c9d1d9;
    white-space: pre-wrap;
    word-break: break-word;
}

.error {
    color: #ef4444;
    background: #2d1b1b;
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid #ef4444;
}
```

- [ ] **Step 2: Commit**

```bash
git add uzinha/static/style.css
git commit -m "feat(uzinha): add dark theme styles"
```

---

### Task 5: JavaScript do Frontend

**Files:**
- Create: `uzinha/static/app.js`

**Interfaces:**
- Consumes: API `/api/minions` e `/api/minion/?name=...`
- Produz: Dashboard dinâmico com auto-refresh

- [ ] **Step 1: Criar app.js**

```javascript
let refreshInterval;

async function fetchMinions() {
    try {
        const resp = await fetch('/api/minions');
        const data = await resp.json();
        renderMinions(data);
    } catch (err) {
        document.getElementById('minions-grid').innerHTML = 
            '<div class="error">Erro ao conectar com Uzinha</div>';
    }
}

function renderMinions(minions) {
    const grid = document.getElementById('minions-grid');
    
    if (minions.length === 0) {
        grid.innerHTML = '<div class="loading">Nenhum minion configurado</div>';
        return;
    }

    grid.innerHTML = minions.map(m => `
        <div class="minion-card ${m.online ? '' : 'offline'}" onclick="showDetail('${m.name}')">
            <div class="minion-header">
                <span class="minion-name">${m.name}</span>
                <span class="status-dot ${m.online ? '' : 'offline'}"></span>
            </div>
            <div class="minion-info">
                <p>Host: ${m.host}</p>
                ${m.agent ? getAgentInfo(m.agent) : ''}
                ${m.error ? `<p class="error">${m.error}</p>` : ''}
            </div>
        </div>
    `).join('');
}

function getAgentInfo(agentData) {
    try {
        const agent = typeof agentData === 'string' ? JSON.parse(agentData) : agentData;
        return `
            <p>Hostname: ${agent.hostname || 'N/A'}</p>
            <p>Version: ${agent.version || 'N/A'}</p>
            <p>Uptime: ${formatUptime(agent.uptime_seconds)}</p>
        `;
    } catch {
        return '';
    }
}

function formatUptime(seconds) {
    if (!seconds) return 'N/A';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    return `${days}d ${hours}h`;
}

async function showDetail(name) {
    document.getElementById('minions-grid').style.display = 'none';
    document.getElementById('minion-detail').classList.remove('hidden');
    
    const content = document.getElementById('detail-content');
    content.innerHTML = '<div class="loading">Carregando dados...</div>';

    try {
        const resp = await fetch(`/api/minion/?name=${encodeURIComponent(name)}`);
        const data = await resp.json();
        renderDetail(data);
    } catch (err) {
        content.innerHTML = `<div class="error">Erro ao buscar dados: ${err.message}</div>`;
    }
}

function renderDetail(data) {
    const content = document.getElementById('detail-content');
    
    let html = `<h2>${data.name}</h2>`;
    
    if (data.error) {
        html += `<div class="error">${data.error}</div>`;
    }

    if (data.agent) {
        html += renderSection('Agent', data.agent);
    }
    if (data.system) {
        html += renderSection('System', data.system);
    }
    if (data.memory) {
        html += renderSection('Memory', data.memory);
    }
    if (data.disk) {
        html += renderSection('Disk', data.disk);
    }
    if (data.users) {
        html += renderSection('Users', data.users);
    }

    html += `<div class="detail-section">
        <h3>Raw JSON</h3>
        <div class="json-view">${JSON.stringify(data, null, 2)}</div>
    </div>`;

    content.innerHTML = html;
}

function renderSection(title, data) {
    let parsed;
    try {
        parsed = typeof data === 'string' ? JSON.parse(data) : data;
    } catch {
        parsed = data;
    }

    return `
        <div class="detail-section">
            <h3>${title}</h3>
            <div class="detail-grid">
                ${Object.entries(parsed).map(([key, val]) => `
                    <div class="detail-item">
                        <label>${key}</label>
                        <span>${typeof val === 'object' ? JSON.stringify(val) : val}</span>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}

function showGrid() {
    document.getElementById('minions-grid').style.display = 'grid';
    document.getElementById('minion-detail').classList.add('hidden');
}

document.addEventListener('DOMContentLoaded', () => {
    fetchMinions();
    refreshInterval = setInterval(fetchMinions, 30000);
});
```

- [ ] **Step 2: Commit**

```bash
git add uzinha/static/app.js
git commit -m "feat(uzinha): add frontend JavaScript with auto-refresh"
```

---

### Task 6: Script de inicialização

**Files:**
- Create: `uzinha/run.sh`

**Interfaces:**
- Consumes: Binário uzinha compilado
- Produz: Servidor rodando

- [ ] **Step 1: Criar run.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ ! -f uzinha ]]; then
  echo "Compilando Uzinha..."
  go build -o uzinha .
fi

echo "Iniciando Uzinha..."
exec ./uzinha
```

- [ ] **Step 2: Tornar executável e commit**

```bash
chmod +x uzinha/run.sh
git add uzinha/run.sh
git commit -m "feat(uzinha): add startup script"
```

---

### Task 7: Documentação

**Files:**
- Create: `docs/uzinha.md`

**Interfaces:**
- Consumes: Todos os arquivos da Uzinha
- Produz: Documentação completa

- [ ] **Step 1: Criar documentação**

```markdown
# Uzinha - Minion Control Panel

A "Uzinha" é um painel de controle web para monitorar e gerenciar múltiplos minions remotamente.

## Início Rápido

```bash
# 1. Configurar minions
vim uzinha/config.json

# 2. Compilar e rodar
cd uzinha
chmod +x run.sh
./run.sh

# 3. Abrir no navegador
open http://localhost:8080
```

## Configuração

Edite `uzinha/config.json`:

```json
{
  "minions": [
    {
      "name": "meu-servidor",
      "host": "https://192.168.1.100:9870",
      "api_key": "minion_sk_...",
      "insecure": true
    }
  ],
  "server": {
    "port": 8080
  }
}
```

### Campos

| Campo | Descrição |
|---|---|
| `name` | Nome identificador do minion |
| `host` | URL completa do minion (https://ip:porta) |
| `api_key` | API key para autenticação |
| `insecure` | Ignorar verificação TLS (true para self-signed) |

## Endpoints da API

A Uzinha expõe:

| Endpoint | Descrição |
|---|---|
| `GET /` | Dashboard principal |
| `GET /api/minions` | Lista todos os minions com dados básicos |
| `GET /api/minion/?name=X` | Dados completos de um minion |

## Dados Coletados

Para cada minion, a Uzinha coleta:

- **Agent**: agent_id, hostname, versão, uptime, capabilities
- **System**: OS, kernel, hostname
- **Memory**: Total, disponível, livre
- **Disk**: Uso de disco por partição
- **Users**: Usuários do sistema

## Funcionalidades

- Dashboard com cards de cada minion
- Indicador online/offline
- Auto-refresh a cada 30 segundos
- Detalhes completos ao clicar em um minion
- Visualização JSON raw
- Tema escuro

## Exemplo de Uso

```bash
# Rodar em background
nohup ./uzinha &

# Verificar minions
curl http://localhost:8080/api/minions

# Ver detalhes de um minion
curl "http://localhost:8080/api/minion/?name=minion-local"
```

## Solução de Problemas

### Minion aparece offline
- Verifique se o minion está rodando: `curl -k https://IP:9870/api/v1/health`
- Verifique a API key no config.json
- Verifique se o IP está correto

### Erro de conexão TLS
- Defina `"insecure": true` no config.json para TLS self-signed

### Dados não atualizam
- O dashboard faz refresh automático a cada 30 segundos
- Recarregue a página manualmente com F5
```

- [ ] **Step 2: Commit**

```bash
git add docs/uzinha.md
git commit -m "docs(uzinha): add usage documentation"
```

---

### Task 8: Atualizar .gitignore

**Files:**
- Modify: `.gitignore`

**Interfaces:**
- Consumes: Binário uzinha compilado
- Produz: Gitignore atualizado

- [ ] **Step 1: Adicionar uzinha ao .gitignore**

```
# Uzinha compiled binary
uzinha/uzinha
```

- [ ] **Step 2: Commit**

```bash
git add .gitignore
git commit -m "chore: add uzinha binary to gitignore"
```
