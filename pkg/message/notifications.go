package message

import (
	"fmt"
	"strings"
)

const localePTBR = "pt-BR"

var notificationTemplates = map[string]string{
	KeyRemindersSlotInMinutes:      "{{slot}} em {{minutes}} min",
	KeyRemindersRecommendedTime:    "Horário recomendado: {{time}}",
	KeyRemindersRuntimeUnavailable: "Runtime indisponível",
	KeyRemindersNotAvailable:       "Notificações não disponíveis",
}

// NormalizeNotificationLocale always returns pt-BR.
func NormalizeNotificationLocale(_ string) string {
	return localePTBR
}

// TranslateNotification renders a notification string for the given locale.
func TranslateNotification(_ string, key string, params map[string]string) string {
	template, ok := notificationTemplates[key]
	if !ok {
		return key
	}
	return interpolate(template, params)
}

func interpolate(template string, params map[string]string) string {
	result := template
	for key, value := range params {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

// ReminderTitle builds a localized reminder notification title.
func ReminderTitle(locale, slotLabel string, leadMinutes int) string {
	return TranslateNotification(locale, KeyRemindersSlotInMinutes, map[string]string{
		"slot":    slotLabel,
		"minutes": fmt.Sprintf("%d", leadMinutes),
	})
}

// ReminderBody builds a localized reminder notification body.
func ReminderBody(locale, time string) string {
	return TranslateNotification(locale, KeyRemindersRecommendedTime, map[string]string{
		"time": time,
	})
}

// SlotLabel returns a localized entry/exit label for notifications.
func SlotLabel(_ string, journeyIndex int, isExit bool) string {
	number := fmt.Sprintf("%d", journeyIndex+1)
	if isExit {
		return fmt.Sprintf("Saída %s", number)
	}
	return fmt.Sprintf("Entrada %s", number)
}
