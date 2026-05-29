document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('loginForm');
    const errorMessage = document.getElementById('errorMessage');
    const loginBtn = document.getElementById('loginBtn');

    loginForm.addEventListener('submit', async (evento) => {
        evento.preventDefault();

        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        loginBtn.innerText = "Autenticando...";
        loginBtn.disabled = true;
        errorMessage.style.display = 'none';

        try {
            const resposta = await fetch('http://localhost:8080/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                })
            });

            const dados = await resposta.json();

            if (dados.sucesso) {
                localStorage.setItem('jennie_token', dados.token);
                alert("Login efetuado com sucesso! Redirecionando...");
                window.location.href = 'dashboard.html'; 
            } else {
                errorMessage.innerText = dados.mensagem;
                errorMessage.style.display = 'block';
                loginBtn.innerText = "Entrar";
                loginBtn.disabled = false;
            }

        } catch (erro) {
            console.error("Erro ao conectar com a API:", erro);
            errorMessage.innerText = "Erro ao conectar com o servidor.";
            errorMessage.style.display = 'block';
            loginBtn.innerText = "Entrar";
            loginBtn.disabled = false;
        }
    });
});