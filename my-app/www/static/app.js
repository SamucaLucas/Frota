// ==========================================
// CONFIGURAÇÃO GLOBAL DO SISTEMA
// ==========================================
const API_URL = "http://192.168.3.38:8082";
window.API_URL = API_URL;

// ==========================================
// MOTOR DO MODO CLARO / ESCURO
// ==========================================
// Pega todos os botões de tema (da Home ou do Menu Lateral)
const toggleButtons = document.querySelectorAll('#theme-toggle, .sb-theme-btn');
const currentTheme = localStorage.getItem('theme');

if (currentTheme === 'light') {
    document.body.classList.add('light-mode');
    toggleButtons.forEach(btn => btn.innerHTML = '🌙 Escuro');
} else {
    toggleButtons.forEach(btn => btn.innerHTML = '☀️ Claro');
}

toggleButtons.forEach(btn => {
    btn.addEventListener('click', () => {
        document.body.classList.toggle('light-mode');
        
        let isLight = document.body.classList.contains('light-mode');
        localStorage.setItem('theme', isLight ? 'light' : 'dark');
        
        toggleButtons.forEach(b => b.innerHTML = isLight ? '🌙 Escuro' : '☀️ Claro');
    });
});

// ==========================================
// MANUTENÇÃO DOS DADOS DO CADASTRO (ANTI-PERDA)
// ==========================================
// Atualizado: Não depende mais do action="/cadastrar" do Go
const inputsParaPersistir = document.querySelectorAll('input[name="nome"], input[name="email"], input[name="whatsapp"]');

if (inputsParaPersistir.length > 0) {
    inputsParaPersistir.forEach(input => {
        const valorSalvo = sessionStorage.getItem('cadastro_' + input.name);
        
        // Só preenche se o campo estiver vazio
        if (valorSalvo && !input.value) {
            input.value = valorSalvo;
        }

        input.addEventListener('input', () => {
            sessionStorage.setItem('cadastro_' + input.name, input.value);
        });
    });
}

// ==========================================
// LÓGICA DO CHECKBOX CUSTOMIZADO
// ==========================================
const labelLembrar = document.querySelector('input[name="lembrar"]');

if (labelLembrar && labelLembrar.parentElement) {
    labelLembrar.parentElement.addEventListener('click', function(e) {
        if(e.target.tagName.toLowerCase() === 'span' && e.target.getAttribute('onclick')) {
            return;
        }
        e.preventDefault(); 
        
        const input = this.querySelector('input[name="lembrar"]');
        const checkIcon = this.querySelector('div span'); 
        
        if (input && checkIcon) {
            input.checked = !input.checked;
            checkIcon.style.display = input.checked ? 'block' : 'none';
        }
    });
}

// ==========================================
// SISTEMA DE INSTALAÇÃO DO PWA (HÍBRIDO)
// ==========================================
function iniciarPWA() {
    // 🚨 TRAVA: Se estiver no aplicativo nativo (Lojas), aborta o banner de instalação!
    if (window.Capacitor && window.Capacitor.isNativePlatform()) {
        return;
    }

    let promptInstalacaoAndroid;
    const pwaBanner = document.getElementById('pwa-banner');
    const btnInstalarAndroid = document.getElementById('btn-instalar-android');
    const instrucaoIOS = document.getElementById('instrucao-ios');

    const isStandalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone;

    if (!isStandalone && pwaBanner) {
        const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
        
        if (isIOS && instrucaoIOS) {
            instrucaoIOS.style.display = 'block';
            setTimeout(() => { pwaBanner.classList.add('show'); }, 2000);
        }
    }

    window.addEventListener('beforeinstallprompt', (e) => {
        e.preventDefault();
        promptInstalacaoAndroid = e;

        if (pwaBanner && btnInstalarAndroid) {
            btnInstalarAndroid.style.display = 'block';
            setTimeout(() => { pwaBanner.classList.add('show'); }, 2000);
        }
    });

    if (btnInstalarAndroid) {
        btnInstalarAndroid.addEventListener('click', async () => {
            if (promptInstalacaoAndroid) {
                promptInstalacaoAndroid.prompt();
                const { outcome } = await promptInstalacaoAndroid.userChoice;
                if (outcome === 'accepted') {
                    fecharBannerPWA();
                }
                promptInstalacaoAndroid = null;
            }
        });
    }
}

function fecharBannerPWA() {
    const pwaBanner = document.getElementById('pwa-banner');
    if(pwaBanner) pwaBanner.classList.remove('show');
}

// Inicia a verificação do PWA automaticamente
iniciarPWA();