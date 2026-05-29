document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('jennie_token');
    
    if (!token) {
        window.location.href = 'login.html';
        return;
    }

    const logoutBtn = document.getElementById('logoutBtn');
    
    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => {
            localStorage.removeItem('jennie_token');
            window.location.href = 'login.html';
        });
    }

    async function BuscarDadosSistema() {
        try {
            const resposta = await fetch("http://localhost:8080/api/status");

            if(resposta.ok) {
                const dados = await resposta.json();

                document.getElementById('cpuTemp').innerText = dados.temperatura;
                document.getElementById('ramUsage').innerText = dados.ram_usada;
                document.getElementById('ramTotal').innerText = `de ${dados.ram_total}`;

                const barraDisco = document.querySelector('.progress-fill');
                const textoDisco = document.querySelector('.card.full-width .metric-detail');

                if(barraDisco && textoDisco) {
                    barraDisco.style.width = dados.disco_porcentagem;
                    textoDisco.innerText = `Uso da Partição Raiz: ${dados.disco_porcentagem}`;
                }
            }
        }
        catch(erro) {
            console.error("Erro ao buscar dados da API: ", erro);
            document.getElementById('cpuTemp').innerText = "Erro";
        }
    }

    let caminhoAtual = "/"

    async function BuscarArquivos() {
        const container = document.getElementById('fileListContainer');
        
        try {
            // Imbutimos na url qual subpasta queremos ler
            const url = `http://localhost:8080/api/arquivos?path=${encodeURIComponent(caminhoAtual)}`;
            const resposta = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': token 
                }
            });

            if(resposta.ok) {
                const arquivos = await resposta.json();
                let html = '';

                // Se não estivermos na raiz, desenha o botão de Voltar
                if (caminhoAtual !== '/') {
                    html += `<button onclick="voltarPasta()" class="btn-acao" style="margin-bottom: 15px; background-color: #6c757d;">⬅ Voltar</button>`;
                }

                if(arquivos.length === 0) {
                    html += '<p style="color: #a0a0a0;">Esta pasta está vazia.</p>';
                    container.innerHTML = html;
                    return;
                }

                // Monta a estrutura da tabela
                html += '<table class="file-table">';
                html += '<tr><th>Nome</th><th>Tamanho</th><th>Modificado em</th><th>Ações</th></tr>';

                // Passa por cada arquivo recebido
                arquivos.forEach(arq => {
                    const icone = arq.is_folder ? '📁' : '📄';
                    const tamanhoFormatado = arq.is_folder ? '-' : (arq.tamanho / 1024 / 1024).toFixed(2) + ' MB';
                    
                    // Se for uma pasta, o nome vira um link clicável chamando a função entrarPasta()
                    const nomeHTML = arq.is_folder 
                        ? `<a href="#" onclick="event.preventDefault(); entrarPasta('${arq.nome}')" style="color: #4CAF50; text-decoration: none; font-weight: bold;">${icone} ${arq.nome}</a>` 
                        : `${icone} ${arq.nome}`;

                    // Se for arquivo mostra o botão Baixar, se for pasta, pode deixar vazio ou colocar botão Abrir
                    const botaoAcao = arq.is_folder 
                        ? `<button class="btn-acao" onclick="event.preventDefault(); entrarPasta('${arq.nome}')" style="background-color: #333;">Abrir</button>` 
                        : `<button class="btn-acao" onclick="baixarArquivo('${arq.nome}')">Baixar</button>`;

                    html += `<tr>
                        <td>${nomeHTML}</td>
                        <td>${tamanhoFormatado}</td>
                        <td>${arq.data}</td>
                        <td>${botaoAcao}</td>
                    </tr>`;
                });

                html += '</table>';
                container.innerHTML = html; 

            } else {
                container.innerHTML = '<p style="color: red;">Erro ao acessar a pasta.</p>';
            }
        }
        catch(erro) {
            console.error("Erro ao buscar arquivos: ", erro);
            container.innerHTML = '<p style="color: red;">Erro de conexão com o servidor.</p>';
        }
    }

    // Funçao para entrar na pasta
    window.entrarPasta = function(nomePasta) {
        if (caminhoAtual === '/') {
            caminhoAtual = '/' + nomePasta;
        } else {
            caminhoAtual = caminhoAtual + '/' + nomePasta;
        }
        BuscarArquivos(); // Recarrega a tabela com o caminho novo
    };

    // Função para voltar da pasta
    window.voltarPasta = function() {
        const partes = caminhoAtual.split('/').filter(p => p !== '');
        partes.pop(); 
        caminhoAtual = '/' + partes.join('/');
        BuscarArquivos(); // Recarrega a tabela com o caminho novo
    };

    window.baixarArquivo = function(nomeArquivo) {
        // Monta o caminho completo do arquivo 
        const caminhoDoArquivo = caminhoAtual === '/' 
            ? '/' + nomeArquivo 
            : caminhoAtual + '/' + nomeArquivo;

        // Cria a URL chamando a nova rota de download, passando o token e o caminho na URL
        const url = `http://localhost:8080/api/download?token=${token}&path=${encodeURIComponent(caminhoDoArquivo)}`;
        
        // Abre a URL, fazendo o navegador iniciar o download nativamente
        window.location.href = url;
    };

    async function realizarUpload(evento) {
        const fileInput = evento.target;
        const arquivoSelecionado = fileInput.files[0];
        
        if (!arquivoSelecionado) {
            return; 
        }

        const formData = new FormData();
        formData.append('arquivo', arquivoSelecionado);

        const btnUpload = document.querySelector('.btn-upload');
        const textoOriginal = btnUpload.innerText;

        try {
            btnUpload.innerText = 'Enviando...';
            btnUpload.disabled = true;

            const url = `http://localhost:8080/api/upload?path=${encodeURIComponent(caminhoAtual)}`;
            const resposta = await fetch(url, {
                method: 'POST',
                headers: {
                    'Authorization': token 
                },
                body: formData
            });

            const dados = await resposta.json();

            if (resposta.ok && dados.sucesso) {
                fileInput.value = ''; 
                BuscarArquivos();    
            } else {
                alert("Erro ao enviar: " + (dados.erro || "Falha desconhecida"));
            }

        } catch (erro) {
            console.error("Erro no upload:", erro);
            alert("Erro de conexão ao tentar enviar o arquivo.");
        } finally {
            btnUpload.innerText = textoOriginal;
            btnUpload.disabled = false;
        }
    }

    const fileInput = document.getElementById('fileUploadInput');
    if (fileInput) {
        fileInput.addEventListener('change', realizarUpload);
    }

    BuscarDadosSistema();
    BuscarArquivos();

    setInterval(BuscarDadosSistema, 3000);
});