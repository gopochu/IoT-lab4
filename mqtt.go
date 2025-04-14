package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client
var device_id = "mqtt-vadlap2002-r7ujy3"

func initMQTTClient(broker string, username string, password string, isCert bool, certName string) {
	// MQTT broker info
	// broker := "tls://dev.rightech.io:8883"
	// username := "vadlap"
	// password := "12345"

	// Загрузка CA сертификата
	var tlsConfig *tls.Config
	if(isCert) {
		certpool := x509.NewCertPool()
		pemCerts, err := ioutil.ReadFile(certName)
		if err != nil {
			log.Fatal(err)
		}
		if ok := certpool.AppendCertsFromPEM(pemCerts); !ok {
			log.Fatal("Не удалось добавить CA сертификат")
		}

		// TLS конфигурация
		tlsConfig = &tls.Config{
			RootCAs: certpool,
		}
	}
	

	// Опции клиента MQTT
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(device_id)
	opts.SetUsername(username)
	opts.SetPassword(password)
	if(isCert) {
		opts.SetTLSConfig(tlsConfig)
	}

	// Создаем клиента MQTT
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	fmt.Println("Подключен к брокеру")
}

func sendData(spots int, topic string) {
	message := strconv.Itoa(spots)
	if client.IsConnected() {
		token := client.Publish(topic, 0, false, message)
		token.Wait()
		fmt.Printf("Опубликовано сообщение: %s в топик: %s\n", message, topic)
	} else {
		fmt.Println("MQTT клиент не подключен")
	}
}

func subscribeToTopic(topic string, messageChannel chan<- string) {
	// Подписываемся на топик
	token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Получено сообщение: %s из топика: %s\n", string(msg.Payload()), msg.Topic())
		messageChannel <- string(msg.Payload()) // Отправляем сообщение в канал
	})

	// Ожидаем завершения подписки
	if token.Wait() && token.Error() != nil {
		fmt.Printf("Ошибка подписки на топик: %v\n", token.Error())
	} else {
		fmt.Printf("Успешно подписан на топик: %s\n", topic)
	}
}