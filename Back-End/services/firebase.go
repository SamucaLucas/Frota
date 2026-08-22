package services

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

// InicializarFirebase tenta carregar as credenciais do arquivo ou de uma Variável de Ambiente (Render).
func InicializarFirebase() {
	var opt option.ClientOption

	// 1. Tenta carregar pela variável de ambiente (Ideal para o Render)
	envJson := os.Getenv("FIREBASE_JSON")
	if envJson != "" {
		opt = option.WithCredentialsJSON([]byte(envJson))
	} else {
		// 2. Se não tem variável, tenta pelo arquivo físico (Local / VPS)
		if _, err := os.Stat("firebase-adminsdk.json"); err == nil {
			opt = option.WithCredentialsFile("firebase-adminsdk.json")
		} else {
			log.Println("⚠️ Aviso: Credenciais do Firebase não encontradas (Nem arquivo firebase-adminsdk.json nem Env FIREBASE_JSON). Push Notifications desativadas.")
			return
		}
	}

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Println("⚠️ Aviso: Arquivo firebase-adminsdk.json não encontrado. Push Notifications nativas desativadas.")
		return
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Println("⚠️ Aviso: Falha ao iniciar FCM Messaging Client:", err)
		return
	}

	fcmClient = client
	log.Println("✅ Firebase Cloud Messaging (Push) ativado com sucesso!")
}

// EnviarPushNotification envia notificação apenas se o FCM estiver configurado e o token for válido.
func EnviarPushNotification(token string, titulo string, corpo string) {
	if fcmClient == nil {
		return // Firebase não configurado, ignora silenciosamente.
	}
	if token == "" {
		return // Usuário não tem token salvo, ignora.
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: titulo,
			Body:  corpo,
		},
		Token: token,
	}

	_, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Println("⚠️ Erro ao enviar Push Notification:", err)
	}
}
