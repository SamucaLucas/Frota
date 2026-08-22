# Plano de Correção: Notificações Push e Conectividade

O aplicativo não está recebendo notificações e provavelmente está com problemas de conexão devido a reversões em arquivos de configuração e erros na implementação do plugin de notificações.

## Análise de Problemas

1. **AndroidManifest.xml Revertido**: O arquivo atual bloqueia tráfego HTTP (`usesCleartextTraffic="false"`) e não referencia a configuração de segurança de rede, o que impede a comunicação com o servidor local e possivelmente com o Firebase.
2. **Capacitor Config Revertido**: As configurações que permitiam o esquema `http` e o plugin `CapacitorHttp` foram removidas, voltando ao padrão `https` que causa erro de "Mixed Content".
3. **Erro no Código JS (`app.js`)**: A variável `capacitorPushNotifications` não existe no escopo global do Capacitor. O correto é acessar via `Capacitor.Plugins.PushNotifications`.
4. **Canais de Notificação**: No Android 8+, as notificações só aparecem se houver um canal (channel) criado. Isso está faltando no JS.

## Mudanças Propostas

### [Capacitor Config] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/capacitor.config.json)

- Restaurar o bloco `server` com `androidScheme: "http"` para evitar bloqueios de segurança em ambiente de teste.
- Reativar o plugin `CapacitorHttp` para facilitar requisições que ignoram CORS.

### [Android Manifest] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/android/app/src/main/AndroidManifest.xml)

- Reativar `android:usesCleartextTraffic="true"`.
- Vincular `android:networkSecurityConfig="@xml/network_security_config"`.
- Adicionar metadados do Firebase para ícone e cor padrão das notificações.

### [Frontend - app.js] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/www/static/app.js)

- Corrigir o acesso ao plugin para `Capacitor.Plugins.PushNotifications`.
- Adicionar a criação de um canal de notificação padrão ("default") logo após o registro.
- Garantir que o Service Worker não seja registrado no ambiente nativo (já que ele bloqueia tráfego HTTP).

## Plano de Verificação

1. **Sincronização**: Executar `npx cap copy android` (solicitar ao usuário).
2. **Build**: Reconstruir o app.
3. **Logs**: Verificar no Logcat se o "FCM Token" é impresso no console.
4. **Teste de Envio**: Enviar uma notificação de teste pelo Console do Firebase.
