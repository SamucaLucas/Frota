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

// --- LÓGICA DO CHECKBOX CUSTOMIZADO (Exclusivo para o Login) ---
const labelLembrar = document.querySelector('input[name="lembrar"]');

if (labelLembrar && labelLembrar.parentElement) {
    labelLembrar.parentElement.addEventListener('click', function(e) {
        // Ignora o clique se o usuário estiver clicando no link "Esqueci minha senha"
        if(e.target.tagName.toLowerCase() === 'span' && e.target.getAttribute('onclick')) {
            return;
        }
        
        e.preventDefault(); // Evita duplo clique no input invisível
        
        const input = this.querySelector('input[name="lembrar"]');
        // Pega especificamente o span que está dentro da div (o nosso "✓")
        const checkIcon = this.querySelector('div span'); 
        
        if (input && checkIcon) {
            input.checked = !input.checked;
            checkIcon.style.display = input.checked ? 'block' : 'none';
        }
    });
}

// --- SISTEMA DE INSTALAÇÃO DO PWA (ANDROID E IOS) ---
let promptInstalacaoAndroid;
const pwaBanner = document.getElementById('pwa-banner');
const btnInstalarAndroid = document.getElementById('btn-instalar-android');
const instrucaoIOS = document.getElementById('instrucao-ios');

// 1. Verifica se já está rodando como App instalado (Standalone)
const isStandalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone;

// Se já estiver instalado, não fazemos nada.
if (!isStandalone && pwaBanner) {
    
    // 2. Detecta se é um dispositivo iOS (iPhone/iPad)
    const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
    
    if (isIOS) {
        // Se for iOS, mostramos o banner com as instruções visuais após 2 segundos
        instrucaoIOS.style.display = 'flex';
        setTimeout(() => {
            pwaBanner.classList.add('show');
        }, 2000);
    }
}

// 3. Captura o evento nativo do Android/Chrome para instalar
window.addEventListener('beforeinstallprompt', (e) => {
    // Previne que o Chrome mostre aquele mini-banner chato automaticamente
    e.preventDefault();
    promptInstalacaoAndroid = e;

    if (pwaBanner && btnInstalarAndroid) {
        // Mostra nosso botão customizado
        btnInstalarAndroid.style.display = 'block';
        
        // Exibe o banner na tela após 2 segundos
        setTimeout(() => {
            pwaBanner.classList.add('show');
        }, 2000);
    }
});

// 4. Lógica do clique no botão de Instalar (Apenas Android/PC)
if (btnInstalarAndroid) {
    btnInstalarAndroid.addEventListener('click', async () => {
        if (promptInstalacaoAndroid) {
            promptInstalacaoAndroid.prompt(); // Mostra o prompt oficial do sistema
            const { outcome } = await promptInstalacaoAndroid.userChoice;
            if (outcome === 'accepted') {
                fecharBannerPWA();
            }
            promptInstalacaoAndroid = null;
        }
    });
}

// Função para fechar o banner
function fecharBannerPWA() {
    if(pwaBanner) pwaBanner.classList.remove('show');
}