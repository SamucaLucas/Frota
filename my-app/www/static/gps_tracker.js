// ==========================================
// MOTOR GLOBAL DE RASTREAMENTO (CAPACITOR)
// ==========================================

window.GPSGlobal = {
    watcherId: null,
    statusAtual: "Disponível", // Status padrão ao logar
    ultimaLat: null,
    ultimaLng: null,
    callbackTela: null,

    obterIdToken: function () {
        const token = localStorage.getItem("token") || sessionStorage.getItem("token");
        if (!token) return null;
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            return payload.usuario_id || payload.id || payload.sub || null;
        } catch (e) {
            return null;
        }
    },

    iniciar: async function () {
        
        if (!window.Capacitor || !window.Capacitor.isNativePlatform()) {
            console.warn("[GPS Global] Rodando no navegador. Background inativo.");
            return;
        }


        if (this.watcherId !== null) return;

        const motoristaId = this.obterIdToken();
        if (!motoristaId) return;

        const baseUrl = window.API_URL || "http://192.168.3.38:8082";

        console.log("[GPS Global] Iniciado para o motorista ID:", motoristaId);
        
        // Garante que o Admin saiba que ele entrou
        this.forcarEnvioStatus("Disponível"); 

        this.watcherId = await window.Capacitor.Plugins.BackgroundGeolocation.addWatcher(
            {
                backgroundMessage: "Sua localização está sendo atualizada.",
                backgroundTitle: "Motorista Ativo",
                requestPermissions: true,
                stale: false,
                distanceFilter: 0
            },
            (location, error) => {
                if (error) return;

                const lat = location.latitude;
                const lng = location.longitude;

                this.ultimaLat = lat;
                this.ultimaLng = lng;

                // 1. Envia para a API do Admin ver no mapa
                fetch(`${baseUrl}/api/motorista/localizacao`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        motorista_id: motoristaId,
                        latitude: lat,
                        longitude: lng,
                        status: this.statusAtual 
                    })
                }).catch(() => { });

                // 2. Se alguma tela estiver plugada (ex: Taxímetro), manda os dados para ela
                if (typeof this.callbackTela === 'function') {
                    this.callbackTela(lat, lng);
                }
            }
        );
    },

    parar: function () {
        // Dispara o status Indisponível antes de desligar os motores
        this.forcarEnvioStatus("Indisponível");

        if (this.watcherId !== null && window.Capacitor) {
            window.Capacitor.Plugins.BackgroundGeolocation.removeWatcher({ id: this.watcherId });
            this.watcherId = null;
        }
    },

    mudarStatus: function (novoStatus) {
        this.statusAtual = novoStatus;
        // Força um envio instantâneo para o Admin ver a mudança na hora, sem esperar o GPS pulsar
        this.forcarEnvioStatus(novoStatus);
    },

    // Função vital para mandar o "Último Suspiro" ou forçar atualizações
    // Função vital para mandar o "Último Suspiro" ou forçar atualizações
    forcarEnvioStatus: function (statusForcado) {
        const motoristaId = this.obterIdToken();
        const token = localStorage.getItem("token") || sessionStorage.getItem("token");
        
        if (!motoristaId) return; 

        const baseUrl = window.API_URL || "http://192.168.3.38:8082";
        const lat = this.ultimaLat || 0.0;
        const lng = this.ultimaLng || 0.0;

        const payload = JSON.stringify({
            motorista_id: parseInt(motoristaId),
            latitude: lat,
            longitude: lng,
            status: statusForcado
        });

        // Usamos o fetch com 'keepalive' em vez de sendBeacon. 
        // Ele garante o envio no fechamento do app e aceita headers de Autenticação!
        fetch(`${baseUrl}/api/motorista/localizacao`, {
            method: 'POST',
            keepalive: true, // A mágica acontece aqui
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}` // Garante que o Go não vai bloquear
            },
            body: payload
        }).catch(err => console.error("Erro no último suspiro:", err));
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

// ==========================================
// GATILHOS DE CICLO DE VIDA DO APP
// ==========================================

// 1. Liga o GPS ao abrir o App (se estiver logado)
document.addEventListener("DOMContentLoaded", () => {
    if (localStorage.getItem("token") || sessionStorage.getItem("token")) {
        window.GPSGlobal.iniciar();
    }
});

// 2. Desliga e avisa o Admin que ficou Indisponível quando o motorista MATA (fecha) o App
window.addEventListener("beforeunload", () => {
    if (window.GPSGlobal.watcherId !== null) {
        window.GPSGlobal.forcarEnvioStatus("Indisponível");
    }
});