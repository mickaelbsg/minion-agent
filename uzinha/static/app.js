let refreshInterval;
let currentTab = 'minions';

function showTab(tab) {
    currentTab = tab;
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.panel').forEach(p => p.classList.add('hidden'));
    
    event.target.classList.add('active');
    document.getElementById(`${tab}-panel`).classList.remove('hidden');
    
    if (tab === 'lxc') {
        refreshLXCList();
    }
}

async function fetchMinions() {
    if (currentTab !== 'minions') return;
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
                ${m.error ? `<p class="error-text">${m.error}</p>` : ''}
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
    } catch { return ''; }
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
    if (data.error) html += `<div class="error">${data.error}</div>`;
    if (data.agent) html += renderSection('Agent', data.agent);
    if (data.system) html += renderSection('System', data.system);
    if (data.memory) html += renderSection('Memory', data.memory);
    if (data.disk) html += renderSection('Disk', data.disk);
    if (data.users) html += renderSection('Users', data.users);
    html += `<div class="detail-section"><h3>Raw JSON</h3><div class="json-view">${JSON.stringify(data, null, 2)}</div></div>`;
    content.innerHTML = html;
}

function renderSection(title, data) {
    let parsed;
    try { parsed = typeof data === 'string' ? JSON.parse(data) : data; } catch { parsed = data; }
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

async function refreshLXCList() {
    const containersDiv = document.getElementById('lxc-containers');
    containersDiv.innerHTML = '<div class="loading">Carregando containers...</div>';
    
    try {
        const resp = await fetch('/api/lxc/list');
        const containers = await resp.json();
        renderLXCContainers(containers);
    } catch (err) {
        containersDiv.innerHTML = `<div class="error">Erro ao listar containers: ${err.message}</div>`;
    }
}

function renderLXCContainers(containers) {
    const div = document.getElementById('lxc-containers');
    if (containers.length === 0) {
        div.innerHTML = '<div class="empty">Nenhum container LXC encontrado</div>';
        return;
    }
    div.innerHTML = containers.map(c => `
        <div class="lxc-card">
            <div class="lxc-header">
                <span class="lxc-name">${c.name}</span>
                <span class="lxc-status ${c.status === 'Running' ? 'running' : 'stopped'}">${c.status}</span>
            </div>
            <div class="lxc-info">
                <p>IP: ${c.IP || 'N/A'}</p>
                <p>Image: ${c.image}</p>
            </div>
            <div class="lxc-actions">
                <button class="btn btn-small btn-primary" onclick="deployMinion('${c.name}')">Deploy Minion</button>
                <button class="btn btn-small btn-secondary" onclick="testMinion('${c.name}')">Testar</button>
                <button class="btn btn-small btn-danger" onclick="destroyContainer('${c.name}')">Destruir</button>
            </div>
        </div>
    `).join('');
}

function showCreateDialog() {
    document.getElementById('create-dialog').classList.remove('hidden');
}

function hideCreateDialog() {
    document.getElementById('create-dialog').classList.add('hidden');
}

async function createContainer(e) {
    e.preventDefault();
    const name = document.getElementById('container-name').value;
    const image = document.getElementById('container-image').value;
    
    addLog(`Criando container ${name}...`);
    
    try {
        const resp = await fetch('/api/lxc/create', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name, image})
        });
        const result = await resp.json();
        if (resp.ok) {
            addLog(`Container ${name} criado com sucesso!`);
            hideCreateDialog();
            refreshLXCList();
        } else {
            addLog(`Erro ao criar container: ${result.error}`);
        }
    } catch (err) {
        addLog(`Erro: ${err.message}`);
    }
}

async function deployMinion(containerName) {
    addLog(`Instalando minion em ${containerName}...`);
    
    try {
        const resp = await fetch('/api/lxc/deploy', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({container_name: containerName})
        });
        const result = await resp.json();
        if (resp.ok) {
            addLog(`Minion instalado em ${containerName}!`);
            if (result.bootstrap) {
                addLog(`Bootstrap credentials: ${result.bootstrap.substring(0, 50)}...`);
            }
            refreshLXCList();
        } else {
            addLog(`Erro ao instalar minion: ${result.error}`);
        }
    } catch (err) {
        addLog(`Erro: ${err.message}`);
    }
}

async function testMinion(containerName) {
    addLog(`Testando minion em ${containerName}...`);
    
    try {
        const resp = await fetch('/api/lxc/test', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name: containerName})
        });
        const results = await resp.json();
        if (resp.ok) {
            results.forEach(r => {
                const icon = r.status === 'pass' ? '✓' : '✗';
                addLog(`${icon} ${r.test}: ${r.output || r.status}`);
            });
        } else {
            addLog(`Erro ao testar: ${results.error}`);
        }
    } catch (err) {
        addLog(`Erro: ${err.message}`);
    }
}

async function destroyContainer(containerName) {
    if (!confirm(`Destruir container ${containerName}?`)) return;
    
    addLog(`Destruindo container ${containerName}...`);
    
    try {
        const resp = await fetch('/api/lxc/destroy', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name: containerName})
        });
        const result = await resp.json();
        if (resp.ok) {
            addLog(`Container ${containerName} destruído!`);
            refreshLXCList();
        } else {
            addLog(`Erro ao destruir: ${result.error}`);
        }
    } catch (err) {
        addLog(`Erro: ${err.message}`);
    }
}

function addLog(message) {
    const logDiv = document.getElementById('log-content');
    const time = new Date().toLocaleTimeString();
    logDiv.innerHTML = `<div class="log-entry">[${time}] ${message}</div>` + logDiv.innerHTML;
}

document.addEventListener('DOMContentLoaded', () => {
    fetchMinions();
    refreshInterval = setInterval(fetchMinions, 30000);
});
