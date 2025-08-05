package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var indexHTML string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mock Server UI</title>
    <style>
        :root {
            /* Светлая тема */
            --bg-color: #f5f5f5;
            --card-bg: white;
            --text-color: #333;
            --border-color: #ddd;
            --input-bg: white;
            --input-border: #ddd;
            --mock-item-bg: #f9f9f9;
            --headers-bg: #e9ecef;
            --shadow: rgba(0,0,0,0.1);
            --success-bg: #d4edda;
            --success-color: #155724;
            --success-border: #c3e6cb;
            --error-bg: #f8d7da;
            --error-color: #721c24;
            --error-border: #f5c6cb;
        }

        [data-theme="dark"] {
            /* Темная тема */
            --bg-color: #1a1a1a;
            --card-bg: #2d2d2d;
            --text-color: #e0e0e0;
            --border-color: #444;
            --input-bg: #3a3a3a;
            --input-border: #555;
            --mock-item-bg: #333;
            --headers-bg: #404040;
            --shadow: rgba(0,0,0,0.3);
            --success-bg: #1e4620;
            --success-color: #a3d9a5;
            --success-border: #2d5a2f;
            --error-bg: #4a1e1e;
            --error-color: #f5a3a3;
            --error-border: #663333;
        }

        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background-color: var(--bg-color); 
            color: var(--text-color);
            transition: background-color 0.3s ease, color 0.3s ease;
        }
        .container { max-width: 1200px; margin: 0 auto; overflow-x: auto; }
        .card { 
            background: var(--card-bg); 
            padding: 20px; 
            margin: 20px 0; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px var(--shadow);
            transition: background-color 0.3s ease;
        }
        h1, h2 { color: var(--text-color); }
        form { display: grid; gap: 15px; }
        label { font-weight: bold; color: var(--text-color); }
        input, select, textarea { 
            padding: 8px; 
            border: 1px solid var(--input-border); 
            border-radius: 4px; 
            font-size: 14px;
            background-color: var(--input-bg);
            color: var(--text-color);
            transition: background-color 0.3s ease, border-color 0.3s ease, color 0.3s ease;
        }
        input:focus, select:focus, textarea:focus {
            outline: none;
            border-color: #007bff;
        }
        textarea { resize: vertical; min-height: 80px; }
        button { 
            padding: 10px 20px; 
            background: #007bff; 
            color: white; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer;
            transition: background-color 0.3s ease;
        }
        button:hover { background: #0056b3; }
        button.delete { background: #dc3545; }
        button.delete:hover { background: #c82333; }
        button.edit { background: #ffc107; color: #000; }
        button.edit:hover { background: #e0a800; }
        .theme-toggle {
            position: fixed;
            top: 20px;
            right: 20px;
            background: #007bff;
            color: white;
            border: 1px solid #007bff;
            border-radius: 20px;
            padding: 8px 12px;
            cursor: pointer;
            font-size: 18px;
            transition: all 0.3s ease;
            z-index: 1000;
        }
        .theme-toggle:hover {
            transform: scale(1.1);
            background: #0056b3;
            border-color: #0056b3;
        }
        [data-theme="dark"] .theme-toggle {
            background: var(--card-bg);
            color: var(--text-color);
            border-color: var(--border-color);
        }
        [data-theme="dark"] .theme-toggle:hover {
            background: var(--headers-bg);
            border-color: var(--border-color);
        }
        .mock-item { 
            border: 1px solid var(--border-color); 
            padding: 15px; 
            margin: 10px 0; 
            border-radius: 4px; 
            background: var(--mock-item-bg); 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            max-width: 100%;
            transition: background-color 0.3s ease, border-color 0.3s ease;
        }
        .mock-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
        .method { padding: 4px 8px; border-radius: 4px; color: white; font-weight: bold; margin-right: 10px; }
        .method.GET { background: #28a745; }
        .method.POST { background: #007bff; }
        .method.PUT { background: #ffc107; color: #000; }
        .method.DELETE { background: #dc3545; }
        .method.PATCH { background: #6c757d; }
        .path { 
            font-family: monospace; 
            font-size: 16px; 
            font-weight: bold; 
            word-break: break-all; 
            overflow-wrap: break-word;
            color: var(--text-color);
        }
        .response-details { 
            margin-top: 10px; 
            font-size: 14px; 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            max-width: 100%; 
        }
        .headers { 
            background: var(--headers-bg); 
            padding: 8px; 
            border-radius: 4px; 
            margin: 5px 0; 
            word-wrap: break-word; 
            word-break: break-all; 
            overflow-wrap: break-word; 
            max-width: 100%;
            transition: background-color 0.3s ease;
        }
        .status-code { font-weight: bold; }
        .message { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { 
            background: var(--success-bg); 
            color: var(--success-color); 
            border: 1px solid var(--success-border);
            transition: background-color 0.3s ease, color 0.3s ease, border-color 0.3s ease;
        }
        .error { 
            background: var(--error-bg); 
            color: var(--error-color); 
            border: 1px solid var(--error-border);
            transition: background-color 0.3s ease, color 0.3s ease, border-color 0.3s ease;
        }
        
        /* Стили для табов */
        .tabs-container {
            margin: 20px 0;
            border-bottom: 2px solid var(--border-color);
        }
        .tabs-header {
            display: flex;
            gap: 0;
            margin-bottom: 0;
        }
        .tab-button {
            padding: 12px 24px;
            background: var(--mock-item-bg);
            border: 1px solid var(--border-color);
            border-bottom: none;
            cursor: pointer;
            border-radius: 8px 8px 0 0;
            color: var(--text-color);
            font-weight: bold;
            transition: all 0.3s ease;
            position: relative;
            z-index: 1;
        }
        .tab-button.active {
            background: var(--card-bg);
            border-bottom: 2px solid var(--card-bg);
            margin-bottom: -2px;
            z-index: 2;
        }
        .tab-button:hover:not(.active) {
            background: var(--headers-bg);
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        
        /* Стили для логов */
        .log-item {
            border: 1px solid var(--border-color);
            padding: 15px;
            margin: 10px 0;
            border-radius: 4px;
            background: var(--mock-item-bg);
            transition: background-color 0.3s ease, border-color 0.3s ease;
        }
        .log-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
            flex-wrap: wrap;
            gap: 10px;
        }
        .log-time {
            font-size: 12px;
            color: #666;
            font-family: monospace;
        }
        [data-theme="dark"] .log-time {
            color: #aaa;
        }
        .duration {
            font-size: 12px;
            padding: 2px 6px;
            border-radius: 3px;
            background: var(--headers-bg);
            color: var(--text-color);
            font-family: monospace;
            font-weight: bold;
        }
        .log-details {
            font-size: 14px;
            margin-top: 10px;
        }
        .log-section {
            margin: 8px 0;
        }
        .log-section-title {
            font-weight: bold;
            margin-bottom: 4px;
            color: var(--text-color);
        }
        .log-data {
            background: var(--headers-bg);
            padding: 8px;
            border-radius: 4px;
            margin: 4px 0;
            word-wrap: break-word;
            overflow-wrap: break-word;
            max-width: 100%;
            font-family: monospace;
            font-size: 12px;
            white-space: pre-wrap;
        }
        .logs-controls {
            margin-bottom: 15px;
            display: flex;
            gap: 10px;
            align-items: center;
            flex-wrap: wrap;
        }
    </style>
</head>
<body>
    <div class="theme-toggle" onclick="toggleTheme()" title="Переключить тему">
        <span id="themeIcon">🌙</span>
    </div>
    <div class="container">
        <h1>🎭 Mock Server UI</h1>
        
        <div id="message"></div>
        
        <!-- Система табов -->
        <div class="tabs-container">
            <div class="tabs-header">
                <div class="tab-button active" onclick="switchTab('mocks')">📋 Управление моками</div>
                <div class="tab-button" onclick="switchTab('logs')">📊 Логи запросов</div>
            </div>
        </div>
        
        <!-- Вкладка управления моками -->
        <div id="mocks-tab" class="tab-content active">
            <div class="card">
                <h2 id="formTitle">Add New Mock</h2>
                <form id="mockForm">
                    <input type="hidden" id="editMode" value="false">
                    <input type="hidden" id="originalPath" value="">
                    <input type="hidden" id="originalMethod" value="">
                    
                    <label for="method">HTTP Method:</label>
                    <select id="method" required>
                        <option value="GET">GET</option>
                        <option value="POST">POST</option>
                        <option value="PUT">PUT</option>
                        <option value="DELETE">DELETE</option>
                        <option value="PATCH">PATCH</option>
                    </select>
                    
                    <label for="path">Path:</label>
                    <input type="text" id="path" placeholder="/api/users" required>
                    
                    <label for="statusCode">Status Code:</label>
                    <input type="number" id="statusCode" value="200" min="100" max="599" required>
                    
                    <label for="headers">Headers (JSON):</label>
                    <textarea id="headers" placeholder='{"Content-Type": "application/json"}'>{}</textarea>
                    
                    <label for="body">Response Body:</label>
                    <textarea id="body" placeholder='{"message": "Hello World"}'></textarea>
                    
                    <button type="submit" id="submitButton">Add Mock</button>
                    <button type="button" id="cancelEdit" onclick="cancelEdit()" style="display: none; background: #6c757d;">Cancel</button>
                </form>
            </div>

            <div class="card">
                <h2>Existing Mocks</h2>
                <div style="margin-bottom: 15px;">
                    <button onclick="loadMocks()">🔄 Refresh List</button>
                    <label style="margin-left: 20px;">
                        <input type="checkbox" id="showFullContent" onchange="loadMocks()"> 
                        Show Full Content
                    </label>
                </div>
                <div id="mocksList"></div>
            </div>
        </div>
        
        <!-- Вкладка логов -->
        <div id="logs-tab" class="tab-content">
            <div class="card">
                <h2>Request Logs</h2>
                <div class="logs-controls">
                    <button onclick="loadLogs()">🔄 Refresh Logs</button>
                    <button onclick="clearLogs()" style="background: #dc3545;">🗑️ Clear Logs</button>
                    <label>
                        <input type="checkbox" id="showFullLogContent" onchange="loadLogs()"> 
                        Show Full Content
                    </label>
                </div>
                <div id="logsList"></div>
            </div>
        </div>
    </div>

    <script>
        // Функции управления темой
        function initTheme() {
            const savedTheme = localStorage.getItem('theme') || 'light';
            const themeIcon = document.getElementById('themeIcon');
            
            if (savedTheme === 'dark') {
                document.documentElement.setAttribute('data-theme', 'dark');
                themeIcon.textContent = '☀️';
            } else {
                document.documentElement.removeAttribute('data-theme');
                themeIcon.textContent = '🌙';
            }
        }

        function toggleTheme() {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const themeIcon = document.getElementById('themeIcon');
            
            if (currentTheme === 'dark') {
                document.documentElement.removeAttribute('data-theme');
                localStorage.setItem('theme', 'light');
                themeIcon.textContent = '🌙';
            } else {
                document.documentElement.setAttribute('data-theme', 'dark');
                localStorage.setItem('theme', 'dark');
                themeIcon.textContent = '☀️';
            }
        }

        function showMessage(text, isError = false) {
            const messageDiv = document.getElementById('message');
            messageDiv.innerHTML = '<div class="message ' + (isError ? 'error' : 'success') + '">' + text + '</div>';
            setTimeout(() => messageDiv.innerHTML = '', 5000);
        }

        document.getElementById('mockForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const isEditMode = document.getElementById('editMode').value === 'true';
            const method = document.getElementById('method').value;
            const path = document.getElementById('path').value;
            const statusCode = parseInt(document.getElementById('statusCode').value);
            const headersText = document.getElementById('headers').value;
            const body = document.getElementById('body').value;

            let headers;
            try {
                headers = JSON.parse(headersText);
            } catch (e) {
                showMessage('Error in headers JSON: ' + e.message, true);
                return;
            }

            const mockData = {
                method: method,
                path: path,
                response: {
                    status_code: statusCode,
                    headers: headers,
                    body: body
                }
            };

            try {
                if (isEditMode) {
                    const originalPath = document.getElementById('originalPath').value;
                    const originalMethod = document.getElementById('originalMethod').value;
                    
                    await fetch('/__mock/delete', {
                        method: 'DELETE',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            method: originalMethod,
                            path: originalPath
                        })
                    });
                }

                const response = await fetch('/__mock/add', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(mockData)
                });

                if (response.ok) {
                    showMessage(isEditMode ? 'Mock successfully updated!' : 'Mock successfully added!');
                    resetForm();
                    loadMocks();
                } else {
                    const error = await response.text();
                    showMessage('Error: ' + error, true);
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, true);
            }
        });

        function resetForm() {
            document.getElementById('mockForm').reset();
            document.getElementById('headers').value = '{}';
            document.getElementById('statusCode').value = '200';
            document.getElementById('editMode').value = 'false';
            document.getElementById('originalPath').value = '';
            document.getElementById('originalMethod').value = '';
            document.getElementById('formTitle').textContent = 'Add New Mock';
            document.getElementById('submitButton').textContent = 'Add Mock';
            document.getElementById('cancelEdit').style.display = 'none';
        }

        function editMock(path, method, mockData) {
            document.getElementById('editMode').value = 'true';
            document.getElementById('originalPath').value = path;
            document.getElementById('originalMethod').value = method;
            document.getElementById('method').value = method;
            document.getElementById('path').value = path;
            document.getElementById('statusCode').value = mockData.status_code;
            document.getElementById('headers').value = JSON.stringify(mockData.headers || {}, null, 2);
            document.getElementById('body').value = mockData.body || '';
            document.getElementById('formTitle').textContent = 'Edit Mock';
            document.getElementById('submitButton').textContent = 'Update Mock';
            document.getElementById('cancelEdit').style.display = 'inline-block';
            
            document.getElementById('mockForm').scrollIntoView({ behavior: 'smooth' });
        }

        function cancelEdit() {
            resetForm();
        }

        async function deleteMock(path, method) {
            if (!confirm('Delete mock ' + method + ' ' + path + '?')) {
                return;
            }

            try {
                const response = await fetch('/__mock/delete', {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        method: method,
                        path: path
                    })
                });

                if (response.ok) {
                    showMessage('Mock deleted!');
                    loadMocks();
                } else {
                    const error = await response.text();
                    showMessage('Error: ' + error, true);
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, true);
            }
        }

        async function loadMocks() {
            try {
                const response = await fetch('/__mock/list');
                if (response.ok) {
                    const mocks = await response.json();
                    displayMocks(mocks);
                } else {
                    showMessage('Error loading mocks', true);
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, true);
            }
        }

        let currentMocks = {};

        function displayMocks(mocks) {
            const mocksList = document.getElementById('mocksList');
            const showFullContent = document.getElementById('showFullContent').checked;
            
            currentMocks = mocks;
            
            if (!mocks || Object.keys(mocks).length === 0) {
                mocksList.innerHTML = '<p>No active mocks</p>';
                return;
            }

            let html = '';
            for (const path in mocks) {
                for (const method in mocks[path]) {
                    const mock = mocks[path][method];
                    const mockId = btoa(path + '|' + method);
                    
                    html += '<div class="mock-item" data-path="' + path + '" data-method="' + method + '">';
                    html += '<div class="mock-header">';
                    html += '<div>';
                    html += '<span class="method ' + method + '">' + method + '</span>';
                    html += '<span class="path">' + path + '</span>';
                    html += '</div>';
                    html += '<div>';
                    html += '<button class="edit" onclick="editMockById(\'' + mockId + '\')" style="margin-right: 10px;">✏️ Edit</button>';
                    html += '<button class="delete" onclick="deleteMock(\'' + path.replace(/'/g, "\\'") + '\', \'' + method + '\')">🗑️ Delete</button>';
                    html += '</div>';
                    html += '</div>';
                    html += '<div class="response-details">';
                    html += '<div><span class="status-code">Status:</span> ' + mock.status_code + '</div>';
                    
                    if (mock.headers && Object.keys(mock.headers).length > 0) {
                        html += '<div><strong>Headers:</strong></div>';
                        const headersJson = JSON.stringify(mock.headers, null, 2);
                        if (showFullContent || headersJson.length <= 200) {
                            html += '<div class="headers">' + headersJson + '</div>';
                        } else {
                            html += '<div class="headers">' + headersJson.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                        }
                    }
                    
                    if (mock.body) {
                        html += '<div><strong>Body:</strong></div>';
                        if (showFullContent || mock.body.length <= 200) {
                            html += '<div class="headers" style="white-space: pre-wrap;">' + mock.body + '</div>';
                        } else {
                            html += '<div class="headers">' + mock.body.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                        }
                    }
                    html += '</div>';
                    html += '</div>';
                }
            }
            mocksList.innerHTML = html;
        }

        function editMockById(mockId) {
            const decoded = atob(mockId).split('|');
            const path = decoded[0];
            const method = decoded[1];
            const mockData = currentMocks[path][method];
            editMock(path, method, mockData);
        }

        // Функции управления табами
        function switchTab(tabName) {
            // Убираем активный класс у всех табов и кнопок
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            document.querySelectorAll('.tab-button').forEach(button => {
                button.classList.remove('active');
            });
            
            // Активируем выбранный таб
            document.getElementById(tabName + '-tab').classList.add('active');
            event.target.classList.add('active');
            
            // Загружаем данные для вкладки логов
            if (tabName === 'logs') {
                loadLogs();
            }
        }

        // Функции работы с логами
        async function loadLogs() {
            try {
                const response = await fetch('/__mock/logs');
                if (response.ok) {
                    const logs = await response.json();
                    displayLogs(logs);
                } else {
                    showMessage('Error loading logs', true);
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, true);
            }
        }

        async function clearLogs() {
            if (!confirm('Clear all request logs?')) {
                return;
            }

            try {
                const response = await fetch('/__mock/logs/clear', {
                    method: 'DELETE'
                });

                if (response.ok) {
                    showMessage('Logs cleared!');
                    loadLogs();
                } else {
                    const error = await response.text();
                    showMessage('Error: ' + error, true);
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, true);
            }
        }

        function displayLogs(logs) {
            const logsList = document.getElementById('logsList');
            const showFullContent = document.getElementById('showFullLogContent').checked;
            
            if (!logs || logs.length === 0) {
                logsList.innerHTML = '<p>No request logs</p>';
                return;
            }

            let html = '';
            for (const log of logs) {
                html += '<div class="log-item">';
                
                // Заголовок лога
                html += '<div class="log-header">';
                html += '<div>';
                html += '<span class="method ' + log.method + '">' + log.method + '</span>';
                html += '<span class="path">' + log.path + '</span>';
                html += '</div>';
                html += '<div>';
                html += '<span class="log-time">' + new Date(log.timestamp).toLocaleString() + '</span>';
                html += '<span class="duration">' + (log.duration / 1000000).toFixed(2) + 'ms</span>';
                html += '</div>';
                html += '</div>';
                
                // Детали лога
                html += '<div class="log-details">';
                html += '<div><span class="status-code">Status:</span> ' + log.status_code + '</div>';
                
                // Заголовки запроса
                if (log.request_headers && Object.keys(log.request_headers).length > 0) {
                    html += '<div class="log-section">';
                    html += '<div class="log-section-title">Request Headers:</div>';
                    const reqHeadersJson = JSON.stringify(log.request_headers, null, 2);
                    if (showFullContent || reqHeadersJson.length <= 200) {
                        html += '<div class="log-data">' + reqHeadersJson + '</div>';
                    } else {
                        html += '<div class="log-data">' + reqHeadersJson.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                    }
                    html += '</div>';
                }
                
                // Тело запроса
                if (log.request_body) {
                    html += '<div class="log-section">';
                    html += '<div class="log-section-title">Request Body:</div>';
                    if (showFullContent || log.request_body.length <= 200) {
                        html += '<div class="log-data">' + log.request_body + '</div>';
                    } else {
                        html += '<div class="log-data">' + log.request_body.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                    }
                    html += '</div>';
                }
                
                // Заголовки ответа
                if (log.response_headers && Object.keys(log.response_headers).length > 0) {
                    html += '<div class="log-section">';
                    html += '<div class="log-section-title">Response Headers:</div>';
                    const respHeadersJson = JSON.stringify(log.response_headers, null, 2);
                    if (showFullContent || respHeadersJson.length <= 200) {
                        html += '<div class="log-data">' + respHeadersJson + '</div>';
                    } else {
                        html += '<div class="log-data">' + respHeadersJson.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                    }
                    html += '</div>';
                }
                
                // Тело ответа
                if (log.response_body) {
                    html += '<div class="log-section">';
                    html += '<div class="log-section-title">Response Body:</div>';
                    if (showFullContent || log.response_body.length <= 200) {
                        html += '<div class="log-data">' + log.response_body + '</div>';
                    } else {
                        html += '<div class="log-data">' + log.response_body.substring(0, 200) + '...<br><small><em>Enable "Show Full Content" to view completely</em></small></div>';
                    }
                    html += '</div>';
                }
                
                html += '</div>';
                html += '</div>';
            }
            
            logsList.innerHTML = html;
        }

        // Инициализация при загрузке страницы
        initTheme();
        loadMocks();
    </script>
</body>
</html>`

type MockResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type MockRoute struct {
	Method   string       `json:"method"`
	Path     string       `json:"path"`
	Response MockResponse `json:"response"`
}

type RequestLog struct {
	ID              int               `json:"id"`
	Timestamp       time.Time         `json:"timestamp"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	RequestHeaders  map[string]string `json:"request_headers"`
	RequestBody     string            `json:"request_body"`
	ResponseHeaders map[string]string `json:"response_headers"`
	ResponseBody    string            `json:"response_body"`
	StatusCode      int               `json:"status_code"`
	Duration        time.Duration     `json:"duration"`
}

var (
	mocks        = make(map[string]map[string]MockResponse) // path -> method -> response
	mu           sync.RWMutex
	requestLogs  []RequestLog
	logsMu       sync.RWMutex
	logIDCounter int
	maxLogs      = 1000 // максимальное количество логов в памяти
	enableTunnel = flag.Bool("tunnel", false, "Enable VK tunnel for external access")
	tunnelShort  = flag.Bool("t", false, "Enable VK tunnel for external access (short form)")
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	if methodMap, ok := mocks[r.URL.Path]; ok {
		if resp, ok := methodMap[r.Method]; ok {
			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(resp.StatusCode)
			w.Write([]byte(resp.Body))
			return
		}
	}

	http.NotFound(w, r)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func addRequestLog(method, path string, reqHeaders map[string]string, reqBody string,
	respHeaders map[string]string, respBody string, statusCode int, duration time.Duration) {
	logsMu.Lock()
	defer logsMu.Unlock()

	logIDCounter++
	newLog := RequestLog{
		ID:              logIDCounter,
		Timestamp:       time.Now(),
		Method:          method,
		Path:            path,
		RequestHeaders:  reqHeaders,
		RequestBody:     reqBody,
		ResponseHeaders: respHeaders,
		ResponseBody:    respBody,
		StatusCode:      statusCode,
		Duration:        duration,
	}

	requestLogs = append(requestLogs, newLog)

	// Ограничиваем количество логов
	if len(requestLogs) > maxLogs {
		requestLogs = requestLogs[len(requestLogs)-maxLogs:]
	}
}

func logRequestMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__mock/ui" || r.URL.Path == "/__mock/list" ||
			r.URL.Path == "/__mock/add" || r.URL.Path == "/__mock/delete" ||
			r.URL.Path == "/__mock/logs" {
			next(w, r)
			return
		}

		startTime := time.Now()

		// Читаем тело запроса
		var reqBody string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				reqBody = string(bodyBytes)
				r.Body = io.NopCloser(strings.NewReader(reqBody))
			}
		}

		reqHeaders := make(map[string]string)
		for name, values := range r.Header {
			if len(values) > 0 {
				reqHeaders[name] = values[0]
			}
		}

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next(rw, r)

		duration := time.Since(startTime)

		respHeaders := make(map[string]string)
		for name, values := range rw.Header() {
			if len(values) > 0 {
				respHeaders[name] = values[0]
			}
		}

		addRequestLog(
			r.Method,
			r.URL.Path,
			reqHeaders,
			reqBody,
			respHeaders,
			string(rw.body),
			rw.statusCode,
			duration,
		)
	}
}

func webUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func listMocksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mocks)
}

func addMockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var route MockRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := mocks[route.Path]; !ok {
		mocks[route.Path] = make(map[string]MockResponse)
	}
	mocks[route.Path][route.Method] = route.Response

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Mock added"))
}

func deleteMockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE allowed", http.StatusMethodNotAllowed)
		return
	}

	var route MockRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if methodMap, ok := mocks[route.Path]; ok {
		delete(methodMap, route.Method)
		if len(methodMap) == 0 {
			delete(mocks, route.Path)
		}
		w.Write([]byte("Mock deleted"))
		return
	}

	http.NotFound(w, r)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	logsMu.RLock()
	defer logsMu.RUnlock()

	reversedLogs := make([]RequestLog, len(requestLogs))
	for i, log := range requestLogs {
		reversedLogs[len(requestLogs)-1-i] = log
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reversedLogs)
}

func clearLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE allowed", http.StatusMethodNotAllowed)
		return
	}

	logsMu.Lock()
	defer logsMu.Unlock()

	requestLogs = []RequestLog{}
	logIDCounter = 0

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logs cleared"))
}

func checkVKTunnelInstalled() bool {
	log.Println("Checking if VK tunnel is installed...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vk-tunnel", "--version")
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("VK tunnel version check timed out (probably means it's installed but hangs)")
		return true
	}

	if err != nil {
		return false
	}

	return true
}

func installVKTunnel() error {
	log.Println("Installing VK tunnel via npm...")
	cmd := exec.Command("npm", "install", "@vkontakte/vk-tunnel", "-g")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	log.Println("VK tunnel installed successfully")
	return nil
}

func openNewTerminal(command string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("osascript", "-e",
			`tell application "Terminal" to do script "`+command+`"`)
		return cmd.Start()
	case "linux":
		terminals := []string{"gnome-terminal", "konsole", "xterm", "x-terminal-emulator"}
		for _, terminal := range terminals {
			if _, err := exec.LookPath(terminal); err == nil {
				var cmd *exec.Cmd
				switch terminal {
				case "gnome-terminal":
					cmd = exec.Command(terminal, "--", "bash", "-c", command+"; read -p 'Press Enter to close...'")
				case "konsole":
					cmd = exec.Command(terminal, "-e", "bash", "-c", command+"; read -p 'Press Enter to close...'")
				default:
					cmd = exec.Command(terminal, "-e", "bash", "-c", command+"; read -p 'Press Enter to close...'")
				}
				return cmd.Start()
			}
		}
		return exec.Command("xterm", "-e", "bash", "-c", command+"; read -p 'Press Enter to close...'").Start()
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", command)
		return cmd.Start()
	default:
		return exec.Command("xterm", "-e", "bash", "-c", command+"; read -p 'Press Enter to close...'").Start()
	}
}

func startVKTunnel() {
	log.Println("Starting VK tunnel...")

	if !checkVKTunnelInstalled() {
		log.Println("VK tunnel not found, installing...")
		if err := installVKTunnel(); err != nil {
			log.Printf("Failed to install VK tunnel: %v", err)
			return
		}
	}

	vkTunnelCmd := "vk-tunnel --insecure=1 --http-protocol=http --ws-protocol=ws --host=localhost --port=8082 --timeout=5000"

	log.Printf("Opening new terminal window for VK tunnel (OS: %s)...", runtime.GOOS)
	log.Println("VK tunnel will run in separate terminal window for interactive authorization")

	if err := openNewTerminal(vkTunnelCmd); err != nil {
		log.Printf("Failed to start VK tunnel in new terminal: %v", err)
		log.Println("Trying fallback method in current terminal...")

		fallbackCmd := exec.Command("sh", "-c", vkTunnelCmd)
		fallbackCmd.Stdin = os.Stdin
		fallbackCmd.Stdout = os.Stdout
		fallbackCmd.Stderr = os.Stderr

		go func() {
			log.Println("Starting VK tunnel in current terminal...")
			log.Println("*** IMPORTANT: You may need to authorize and press ENTER when prompted ***")
			if err := fallbackCmd.Run(); err != nil {
				log.Printf("VK tunnel finished with error: %v", err)
			}
		}()
		return
	}

	log.Println("VK tunnel started in new terminal window")
	log.Println("*** Please complete authorization in the new terminal window ***")
	log.Println("*** After authorization, VK tunnel URLs will appear in the new terminal ***")
}

func main() {
	flag.Parse()

	shouldStartTunnel := *enableTunnel || *tunnelShort

	http.HandleFunc("/__mock/ui", webUIHandler)
	http.HandleFunc("/__mock/list", listMocksHandler)
	http.HandleFunc("/__mock/add", addMockHandler)
	http.HandleFunc("/__mock/delete", deleteMockHandler)
	http.HandleFunc("/__mock/logs", logsHandler)
	http.HandleFunc("/__mock/logs/clear", clearLogsHandler)
	http.HandleFunc("/", logRequestMiddleware(mockHandler))

	log.Println("Dynamic mock server running on :8082")
	log.Println("Web UI available at: http://localhost:8082/__mock/ui")

	if shouldStartTunnel {
		log.Println("VK tunnel mode enabled - external access will be available shortly")
		startVKTunnel()
	}

	log.Println("Starting HTTP server...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}
