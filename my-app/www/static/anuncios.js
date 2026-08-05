// ==========================================
// MOTOR GLOBAL DE ANÚNCIOS (CARROSSEL)
// ==========================================
let listaAnunciosGlobal = [];
let indiceAnuncioAtual = 0;
let idAnuncioAtual = 0;

async function iniciarCarrosselAnuncios() {
    const tokenAd = sessionStorage.getItem("token") || localStorage.getItem("token");
    if (!tokenAd) return;

    try {
        const res = await fetch(`${window.API_URL}/api/anuncios/ativos`, {
            headers: { "Authorization": `Bearer ${tokenAd}` }
        });
        
        if (res.ok) {
            listaAnunciosGlobal = await res.json();
            
            if (listaAnunciosGlobal && listaAnunciosGlobal.length > 0) {
                const container = document.getElementById('ad-container');
                if (container) {
                    container.style.display = 'block';
                    mostrarAnuncioUI(0);
                    
                    if (listaAnunciosGlobal.length > 1) {
                        setInterval(() => {
                            indiceAnuncioAtual = (indiceAnuncioAtual + 1) % listaAnunciosGlobal.length;
                            mostrarAnuncioUI(indiceAnuncioAtual);
                        }, 7000); // Rotaciona a cada 7 segundos
                    }
                }
            }
        }
    } catch (err) {
        console.error("Silêncio: Falha ao carregar anúncios.");
    }
}

function mostrarAnuncioUI(index) {
    const ad = listaAnunciosGlobal[index];
    if (!ad) return;

    idAnuncioAtual = ad.id;
    const imgEl = document.getElementById('ad-image');
    const linkEl = document.getElementById('ad-link');

    if (!imgEl || !linkEl) return;

    let imgUrl = ad.imagem_url;
    if (!imgUrl.startsWith("http")) {
        imgUrl = `${window.API_URL}${imgUrl}`;
    }

    imgEl.style.opacity = 0;
    setTimeout(() => {
        imgEl.src = imgUrl;
        linkEl.href = ad.link_destino;
        imgEl.style.opacity = 1;
    }, 400); 
}

function registrarCliqueAd(event) {
    const tokenAd = sessionStorage.getItem("token") || localStorage.getItem("token");
    if (idAnuncioAtual > 0 && tokenAd) {
        fetch(`${window.API_URL}/api/anuncios/${idAnuncioAtual}/clique`, {
            method: 'POST',
            headers: { "Authorization": `Bearer ${tokenAd}` }
        }).catch(() => {});
    }
}

// Auto-inicializa o carrossel se a tela tiver o contêiner de anúncios
document.addEventListener("DOMContentLoaded", () => {
    if (document.getElementById('ad-container')) {
        iniciarCarrosselAnuncios();
    }
});