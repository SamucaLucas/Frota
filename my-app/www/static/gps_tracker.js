// ==========================================
// MOTOR GLOBAL DE RASTREAMENTO (HÍBRIDO: NATIVO E WEB)
// ==========================================

window.GPSGlobal = {
    watcherId: null,
    statusAtual: "Disponível", 
    ultimaLat: null,
    ultimaLng: null,
    callbackTela: null,

    obterIdToken: function () {
        const token = localStorage.getItem("token") || sessionStorage.getItem("token");
        if (!token) return null;
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            return payload.id || payload.ID || payload.Id || payload.usuario_id || payload.UsuarioID || payload.sub || null;
        } catch (e) {
            return null;
        }
    },

    obterToken: function() {
        return localStorage.getItem("token") || sessionStorage.getItem("token") || "";
    },

    iniciar: async function () {
        if (this.watcherId !== null) return;

        const motoristaId = this.obterIdToken();
        if (!motoristaId) return;

        const baseUrl = window.API_URL || "https://sua-api.onrender.com";
        const token = this.obterToken();

        console.log("[GPS Global] Iniciado para o motorista ID:", motoristaId);
        this.forcarEnvioStatus("Disponível"); 

        // 1. SE FOR APLICATIVO NATIVO (ANDROID)
        if (window.Capacitor && window.Capacitor.isNativePlatform()) {
            this.watcherId = await window.Capacitor.Plugins.BackgroundGeolocation.addWatcher(
                {
                    backgroundMessage: "Sua localização está sendo atualizada.",
                    backgroundTitle: "Motorista Ativo",
                    requestPermissions: true,
                    stale: false,
                    distanceFilter: 0
                },
                (location, error) => {
                    if (error) {
                        console.error("Erro no GPS Nativo:", error);
                        return;
                    }
                    this.processarCoordenada(location.latitude, location.longitude, motoristaId, baseUrl, token);
                }
            );
        } 
        // 2. SE FOR NAVEGADOR WEB (PC OU PWA)
        else {
            console.warn("[GPS Global] Rodando no navegador. Usando GPS Web (Foreground).");
            if ("geolocation" in navigator) {
                this.watcherId = navigator.geolocation.watchPosition(
                    (position) => {
                        this.processarCoordenada(position.coords.latitude, position.coords.longitude, motoristaId, baseUrl, token);
                    },
                    (error) => {
                        console.error("Erro no GPS Web:", error);
                    },
                    { enableHighAccuracy: true, maximumAge: 0 }
                );
            } else {
                console.error("Geolocalização não é suportada neste navegador.");
            }
        }
    },

    processarCoordenada: function(lat, lng, motoristaId, baseUrl, token) {
        this.ultimaLat = lat;
        this.ultimaLng = lng;

        // Envia para o Go com o Token de Autorização!
        fetch(`${baseUrl}/api/motorista/localizacao`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}` 
            },
            body: JSON.stringify({
                motorista_id: parseInt(motoristaId),
                latitude: lat,
                longitude: lng,
                status: this.statusAtual 
            })
        }).catch(() => { });

        // Manda os dados para o Taxímetro, se ele estiver aberto
        if (typeof this.callbackTela === 'function') {
            this.callbackTela(lat, lng);
        }
    },

    parar: function () {
        this.forcarEnvioStatus("Indisponível");

        if (this.watcherId !== null) {
            // Desliga o motor correto dependendo de onde o app está rodando
            if (window.Capacitor && window.Capacitor.isNativePlatform()) {
                window.Capacitor.Plugins.BackgroundGeolocation.removeWatcher({ id: this.watcherId });
            } else if ("geolocation" in navigator) {
                navigator.geolocation.clearWatch(this.watcherId);
            }
            this.watcherId = null;
        }
    },

    mudarStatus: function (novoStatus) {
        this.statusAtual = novoStatus;
        this.forcarEnvioStatus(novoStatus);
    },

    forcarEnvioStatus: function (statusForcado) {
        const motoristaId = this.obterIdToken();
        const token = this.obterToken();
        if (!motoristaId) return; 

        const baseUrl = window.API_URL || "https://sua-api.onrender.com";
        const lat = this.ultimaLat || 0.0;
        const lng = this.ultimaLng || 0.0;

        const payload = JSON.stringify({
            motorista_id: parseInt(motoristaId),
            latitude: lat,
            longitude: lng,
            status: statusForcado
        });

        fetch(`${baseUrl}/api/motorista/localizacao`, {
            method: 'POST',
            keepalive: true,
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}` 
            },
            body: payload
        }).catch(err => console.error("Erro no envio forçado:", err));
    },

    calcularDistancia: function (lat1, lon1, lat2, lon2) {
        const R = 6371;
        const dLat = (lat2 - lat1) * Math.PI / 180;
        const dLon = (lon2 - lon1) * Math.PI / 180;
        const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) + Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * Math.sin(dLon / 2) * Math.sin(dLon / 2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        return R * c;
    }
};

document.addEventListener("DOMContentLoaded", () => {
    if (localStorage.getItem("token") || sessionStorage.getItem("token")) {
        window.GPSGlobal.iniciar();
    }
});

window.addEventListener("beforeunload", () => {
    if (window.GPSGlobal.watcherId !== null) {
        window.GPSGlobal.forcarEnvioStatus("Indisponível");
    }
});