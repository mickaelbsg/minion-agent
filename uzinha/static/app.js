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
