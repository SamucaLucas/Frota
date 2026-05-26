// --- MOTOR DO MODO CLARO / ESCURO ---
const toggleButton = document.getElementById('theme-toggle');
const currentTheme = localStorage.getItem('theme');

// Ao carregar a página, verifica se o usuário gosta de modo claro
if (currentTheme === 'light') {
    document.body.classList.add('light-mode');
    if(toggleButton) toggleButton.innerHTML = '🌙 Escuro';
} else {
    // Se for null ou 'dark', garante que o botão mostre "Claro"
    if(toggleButton) toggleButton.innerHTML = '☀️ Claro';
}

// Quando clicar no botão do topo
if(toggleButton) {
    toggleButton.addEventListener('click', () => {
        document.body.classList.toggle('light-mode');
        
        let isLight = document.body.classList.contains('light-mode');
        localStorage.setItem('theme', isLight ? 'light' : 'dark');
        
        toggleButton.innerHTML = isLight ? '🌙 Escuro' : '☀️ Claro';
    });
}

// --- MANUTENÇÃO DOS DADOS DO CADASTRO (ANTI-PERDA DE DADOS) ---
const cadastroForm = document.querySelector('form[action="/cadastrar"]');

if (cadastroForm) {
    // Lista de inputs que desejamos salvar (ignorando senhas e checkbox)
    const inputsParaPersistir = cadastroForm.querySelectorAll('input[name="nome"], input[name="email"], input[name="whatsapp"]');

    // 1. Ao carregar a tela de cadastro, verifica se existem dados guardados na sessão e os restaura
    inputsParaPersistir.forEach(input => {
        const valorSalvo = sessionStorage.getItem('cadastro_' + input.name);
        if (valorSalvo) {
            input.value = valorSalvo;
        }

        // 2. Sempre que o usuário digitar qualquer caractere, salva imediatamente na sessão
        input.addEventListener('input', () => {
            sessionStorage.setItem('cadastro_' + input.name, input.value);
        });
    });

    // 3. Limpa a memória temporária SOMENTE quando o cadastro for feito com sucesso
    cadastroForm.addEventListener('submit', () => {
        // Deixamos os dados salvos temporariamente. Se o servidor aceitar, o fluxo muda.
        // Se der erro (ex: e-mail já cadastrado), o valor continua preenchido para ele corrigir.
    });
}

// Adicione isso ao seu static/app.js
document.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
    checkbox.parentElement.addEventListener('click', function(e) {
        // Se for o label de "Lembrar meus dados"
        if (this.querySelector('input[type="checkbox"]')) {
            const span = this.querySelector('span'); // O elemento do ícone ✓
            const input = this.querySelector('input');
            
            input.checked = !input.checked;
            span.style.display = input.checked ? 'block' : 'none';
        }
    });
});