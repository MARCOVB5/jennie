// Configuração base para facilitar 
const CONFIG = {
    baseUrl: "http://localhost:8080/api",
    token: localStorage.getItem('jennie_token')
}

// Armazena o caminho do diretório
const estado = {
    caminhoAtual: "/"
}

if(!CONFIG.token) {
    window.location.href = 'login.html';
}

// Funções que centraliza as requisições
async function requisicaoAPI(rota, metodo = 'GET', corpo = null) {
    const opcoes = {
        method: metodo,
        headers: { 'Authorization': CONFIG.token }
    };
    
    if (corpo) opcoes.body = corpo;

    try {
        const resposta = await fetch(`${CONFIG.baseUrl}${rota}`, opcoes);
        const dados = await resposta.json();
        return { ok: resposta.ok, dados: dados };
    } catch (erro) {
        console.error(`Erro de rede na rota ${rota}:`, erro);
        return { ok: false, dados: { erro: "Falha de conexão com o servidor." } };
    }
}

// Método para navegar para frente através de pastas nos diretórios
function entrarPasta(nomePasta) {
    estado.caminhoAtual = estado.caminhoAtual === '/' 
        ? '/' + nomePasta 
        : estado.caminhoAtual + '/' + nomePasta;
    atualizarGerenciador();
}

// Método para navegar para trás através de pastas nos diretórios
function voltarPasta() {
    const partes = estado.caminhoAtual.split('/').filter(p => p !== '');
    partes.pop(); 
    estado.caminhoAtual = '/' + partes.join('/');
    atualizarGerenciador();
}

// Método para navegar diretamente para um caminho especifiico
function irParaCaminhoAbsoluto(novoCaminho) {
    estado.caminhoAtual = novoCaminho;
    atualizarGerenciador();
}

function renderizarCaminhoDiretorio() {
    const container = document.getElementById('breadcrumbs');
    
    // O botão inicial Raiz
    let html = `<span style="cursor:pointer; color:#007BFF; font-weight:bold;" onclick="irParaCaminhoAbsoluto('/')">🏠 Início</span>`;
    
    if (estado.caminhoAtual !== '/') {
        const partes = estado.caminhoAtual.split('/').filter(p => p !== '');
        let caminhoAcumulado = '';
        
        partes.forEach((parte, index) => {
            caminhoAcumulado += '/' + parte;
            
            // Adiciona o separador visual
            html += ` <span style="color:#6c757d; font-size: 14px;"> ❯ </span> `;
            
            // Se for a última pasta, fica apenas a texto (sem clique)
            if (index === partes.length - 1) {
                html += `<span style="font-weight:bold; color:#333;">${parte}</span>`;
            } else {
                // Se for uma pasta anterior, torna-se um link clicável
                html += `<span style="cursor:pointer; color:#007BFF;" onclick="irParaCaminhoAbsoluto('${caminhoAcumulado}')">${parte}</span>`;
            }
        });
    }
    
    container.innerHTML = html;
}

// Método para baixar arquivos
function baixarArquivo(nomeArquivo) {
    const caminhoDoArquivo = estado.caminhoAtual === '/' 
        ? '/' + nomeArquivo 
        : estado.caminhoAtual + '/' + nomeArquivo;
    
    // O download do navegador precisa da URL direta
    window.location.href = `${CONFIG.baseUrl}/download?token=${CONFIG.token}&path=${encodeURIComponent(caminhoDoArquivo)}`;
}

// Método para excluir arquivo
async function excluirArquivo(nomeArquivo) {
    // Confirmação do usuário se ele tem certeza que quer excluir
    if (!confirm(`Tem certeza que deseja excluir "${nomeArquivo}"?`)) return;

    const caminho = estado.caminhoAtual === '/' ? '/' + nomeArquivo : estado.caminhoAtual + '/' + nomeArquivo;
    
    const req = await requisicaoAPI(`/excluir?path=${encodeURIComponent(caminho)}`, 'DELETE');
    
    if (req.ok && req.dados.sucesso) {
        atualizarGerenciador();
    } else {
        alert("Erro ao excluir: " + (req.dados.erro || "Falha desconhecida"));
    }
}

// Método para renomearArquivo ou pasta
async function renomearArquivo(nomeAtual) {
    const novoNome = prompt(`Digite o novo nome para "${nomeAtual}":`, nomeAtual);
    
    if (!novoNome || novoNome.trim() === "" || novoNome === nomeAtual) return; 

    const caminho = estado.caminhoAtual === '/' ? '/' + nomeAtual : estado.caminhoAtual + '/' + nomeAtual;
    
    const req = await requisicaoAPI(`/renomear?path=${encodeURIComponent(caminho)}&novoNome=${encodeURIComponent(novoNome.trim())}`, 'POST');
    
    if (req.ok && req.dados.sucesso) {
        atualizarGerenciador();
    } else {
        alert("Erro ao renomear: " + (req.dados.erro || "Falha desconhecida"));
    }
}

// Método para criar nova pasta e salvar no servidor
async function criarNovaPasta() {
    const nomePasta = prompt("Digite o nome da nova pasta:");
    if (!nomePasta || nomePasta.trim() === "") return; 

    const caminho = estado.caminhoAtual === '/' ? '/' + nomePasta.trim() : estado.caminhoAtual + '/' + nomePasta.trim();
    
    const req = await requisicaoAPI(`/criarpasta?path=${encodeURIComponent(caminho)}`, 'POST');
    
    if (req.ok && req.dados.sucesso) {
        atualizarGerenciador();
    } else {
        alert("Erro ao criar pasta: " + (req.dados.erro || "Falha desconhecida"));
    }
}

// Método para fazer o upload de arquivos
async function uploadArquivos(eventoOuArquivo) {
    const arquivo = eventoOuArquivo.target ? eventoOuArquivo.target.files[0] : eventoOuArquivo;
    if (!arquivo) return; 

    const formData = new FormData();
    formData.append('arquivo', arquivo);

    const btnUpload = document.querySelector('.btn-upload');
    const containerProgresso = document.getElementById('containerProgresso');
    const barraProgresso = document.getElementById('barraProgresso');
    const textoProgresso = document.getElementById('textoProgresso');

    const textoOriginal = btnUpload.innerText;
    btnUpload.innerText = 'Enviando...';
    btnUpload.disabled = true;

    // Barra de progresso
    containerProgresso.style.display = 'block';
    barraProgresso.style.width = '0%';
    textoProgresso.innerText = '0%';

    // Usando XMLHttpRequest porque o fetch não mede progresso de upload
    const xhr = new XMLHttpRequest();
    xhr.open('POST', `${CONFIG.baseUrl}/upload?path=${encodeURIComponent(estado.caminhoAtual)}`);
    xhr.setRequestHeader('Authorization', CONFIG.token);

    // Atualiza a barra enquanto envia
    xhr.upload.onprogress = function(event) {
        if (event.lengthComputable) {
            const percentual = Math.round((event.loaded / event.total) * 100);
            barraProgresso.style.width = percentual + '%';
            textoProgresso.innerText = percentual + '%';
        }
    };

    // Quando terminar o envio
    xhr.onload = function() {
        if (xhr.status === 200) {
            if (eventoOuArquivo.target) {
                eventoOuArquivo.target.value = ''; 
            }
            // Atualiza a tabela na tela
            atualizarGerenciador();    
        } else {
            alert("Erro ao enviar. Código: " + xhr.status);
        }

        btnUpload.innerText = textoOriginal;
        btnUpload.disabled = false;
        
        // Mantém a barra em 100% e depois esconde
        setTimeout(() => {
            containerProgresso.style.display = 'none';
        }, 1000);
    };

    // Se der erro de internet
    xhr.onerror = function() {
        alert("Erro de conexão ao tentar enviar o arquivo.");
        btnUpload.innerText = textoOriginal;
        btnUpload.disabled = false;
        containerProgresso.style.display = 'none';
    };

    xhr.send(formData);
}

// Função para clicar no arquivo e abrir a imagem no modal
function abrirImagem(nomeArquivo) {
    const modal = document.getElementById('meuModal');
    const imagem = document.getElementById('imagemAmpliada');
    
    const caminhoDoArquivo = estado.caminhoAtual === '/' ? '/' + nomeArquivo : estado.caminhoAtual + '/' + nomeArquivo;
    
    // Constrói a URL para a nossa nova rota de visualização
    const url = `${CONFIG.baseUrl}/visualizar?token=${CONFIG.token}&path=${encodeURIComponent(caminhoDoArquivo)}`;
    
    imagem.src = url;
    modal.style.display = "block";
}

// Método para fechar a visualização da imagem ao clicar no x
function fecharVisualizador() {
    const modal = document.getElementById('meuModal');
    modal.style.display = "none";
    document.getElementById('imagemAmpliada').src = ""; // Limpa a memória
}

// Método para atualizar a lista de arquivos
async function atualizarGerenciador() {
    const container = document.getElementById('fileListContainer');
    
    renderizarCaminhoDiretorio();

    const req = await requisicaoAPI(`/arquivos?path=${encodeURIComponent(estado.caminhoAtual)}`);

    if(!req.ok) {
        container.innerHTML = '<p style="color: red;">Erro ao acessar a pasta.</p>';
        return;
    }

    renderizarTabelaHTML(req.dados, container);
}

// Método para centralizar a renderização da tabela HTML
function renderizarTabelaHTML(arquivos, container) {
    let html = '';

    if (estado.caminhoAtual !== '/') {
        html += `<button onclick="voltarPasta()" class="btn-acao" style="margin-top: 20px; margin-bottom: 15px; background-color: #6c757d;">⬅ Voltar</button>`;
    }

    if(arquivos.length === 0) {
        html += '<p style="color: #a0a0a0; margin-top: 15px;">Esta pasta está vazia.</p>';
        container.innerHTML = html;
        return;
    }

    html += '<table class="file-table">';
    html += '<tr><th>Nome</th><th>Tamanho</th><th>Modificado em</th><th>Ações</th></tr>';

    arquivos.forEach(arq => {
        const icone = arq.is_folder ? '📁' : '📄';
        const tamanhoFormatado = arq.is_folder ? '-' : (arq.tamanho / 1024 / 1024).toFixed(2) + ' MB';
        
        // Lógica para resumir nomes longos para não quebrar a tabela
        const limiteCaracteres = 25; 
        const nomeExibicao = arq.nome.length > limiteCaracteres 
            ? arq.nome.substring(0, limiteCaracteres) + '...' 
            : arq.nome;
        
        const nomeHTML = arq.is_folder 
            ? `<a href="#" onclick="event.preventDefault(); entrarPasta('${arq.nome}')" title="${arq.nome}" style="color: #4CAF50; text-decoration: none; font-weight: bold;">${icone} ${nomeExibicao}</a>` 
            : `<span title="${arq.nome}">${icone} ${nomeExibicao}</span>`;

        // Detecta se é uma imagem pela extensão
        const extensao = arq.nome.split('.').pop().toLowerCase();
        const imagensExt = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];
        const eImagem = !arq.is_folder && imagensExt.includes(extensao);

        let botaoAcao = "";

        if (arq.is_folder) {
            // Caso se for uma pasta
            botaoAcao = `<button class="btn-acao" onclick="event.preventDefault(); entrarPasta('${arq.nome}')" style="background-color: #333; margin-right: 5px;">Abrir</button>
                         <button class="btn-acao" onclick="renomearArquivo('${arq.nome}')" style="background-color: #17a2b8; color: white; margin-right: 5px;">Renomear</button>
                         <button class="btn-acao" onclick="excluirArquivo('${arq.nome}')" style="background-color: #dc3545;">Excluir</button>`;
        } else {
            // Caso se for uma imagem adiciona um botão
            const botaoVer = eImagem ? `<button class="btn-acao" onclick="abrirImagem('${arq.nome}')" style="background-color: #28a745; color: white; margin-right: 5px;">Ver Foto</button>` : "";
            
            botaoAcao = `${botaoVer}
                         <button class="btn-acao" onclick="baixarArquivo('${arq.nome}')" style="margin-right: 5px;">Baixar</button>
                         <button class="btn-acao" onclick="renomearArquivo('${arq.nome}')" style="background-color: #17a2b8; color: white; margin-right: 5px;">Renomear</button>
                         <button class="btn-acao" onclick="excluirArquivo('${arq.nome}')" style="background-color: #dc3545;">Excluir</button>`;
        }

        html += `<tr><td>${nomeHTML}</td><td>${tamanhoFormatado}</td><td>${arq.data}</td><td>${botaoAcao}</td></tr>`;
    });

    html += '</table>';
    container.innerHTML = html; 
}

// Método que busca as informações de hardware
async function buscarDadosSistema() {
    const req = await requisicaoAPI('/status');
    if(req.ok) {
        document.getElementById('cpuTemp').innerText = req.dados.temperatura;
        document.getElementById('ramUsage').innerText = req.dados.ram_usada;
        document.getElementById('ramTotal').innerText = `de ${req.dados.ram_total}`;

        const barraDisco = document.querySelector('.progress-fill');
        const textoDisco = document.querySelector('.card.full-width .metric-detail');
        if(barraDisco && textoDisco) {
            barraDisco.style.width = req.dados.disco_porcentagem;
            textoDisco.innerText = `Uso da Partição Raiz: ${req.dados.disco_porcentagem}`;
        }
    }
}

// Método para implementar arrastar o arquivo na interface e fazer o upload
function configurarDragAndDrop() {
    // Selecionamos o painel inteiro da tela para ser possível arrastar o arquivo
    const dropZone = document.querySelectorAll('.card.full-width')[1];
    
    if (!dropZone) return;

    // Prevenir o comportamento padrão do navegador, que é abrir o arquivo
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(nomeEvento => {
        dropZone.addEventListener(nomeEvento, prevenirPadrao, false);
    });

    function prevenirPadrao(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    // Adicionar o efeito visual (tracejado) quando o arquivo entra na zona
    ['dragenter', 'dragover'].forEach(nomeEvento => {
        dropZone.addEventListener(nomeEvento, () => {
            dropZone.classList.add('area-arrasto-ativa');
        }, false);
    });

    // Remover o efeito visual quando o arquivo sai da zona ou é largado
    ['dragleave', 'drop'].forEach(nomeEvento => {
        dropZone.addEventListener(nomeEvento, () => {
            dropZone.classList.remove('area-arrasto-ativa');
        }, false);
    });

    // Capturar o arquivo no momento exato em que ele é largado
    dropZone.addEventListener('drop', (e) => {
        const arquivosArrastados = e.dataTransfer.files;
        
        if (arquivosArrastados.length > 0) {
            // Pegamos no primeiro arquivo largado e mandamos para a nossa função de upload
            uploadArquivos(arquivosArrastados[0]); 
        }
    }, false);
}

// Inicialização da tela
document.addEventListener('DOMContentLoaded', () => {
    // Liga o sistema de upload ao botão invisível
    const fileInput = document.getElementById('fileUploadInput');
    if (fileInput) fileInput.addEventListener('change', uploadArquivos);

    // Configura o logout
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) logoutBtn.addEventListener('click', () => {
        localStorage.removeItem('jennie_token');
        window.location.href = 'login.html';
    });

    // Se clicar fora da tela do model de visualização de imagem, fecha o modal
    window.addEventListener('click', (evento) => {
        const modal = document.getElementById('meuModal');
        if (evento.target === modal) {
            fecharVisualizador();
        }
    });

    // Inicia inicial nas leituras da tela
    buscarDadosSistema();
    atualizarGerenciador();
    configurarDragAndDrop();
    setInterval(buscarDadosSistema, 3000);
});