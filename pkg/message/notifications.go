package message

import (
	"fmt"
	"strings"
)

const (
	localeEN   = "en"
	localePTBR = "pt-BR"
)

var notificationTemplates = map[string]map[string]string{
	localeEN: {
		KeyRemindersSlotInMinutes:      "{{slot}} in {{minutes}} min",
		KeyRemindersRecommendedTime:    "Recommended time: {{time}}",
		KeyRemindersRuntimeUnavailable: "Runtime unavailable",
		KeyRemindersNotAvailable:       "Notifications unavailable",
	},
	localePTBR: {
		KeyRemindersSlotInMinutes:      "{{slot}} em {{minutes}} min",
		KeyRemindersRecommendedTime:    "Horário recomendado: {{time}}",
		KeyRemindersRuntimeUnavailable: "Runtime indisponível",
		KeyRemindersNotAvailable:       "Notificações não disponíveis",
	},
}

// NormalizeNotificationLocale maps a locale string to a supported notification locale.
func NormalizeNotificationLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(locale, "en") {
		return localeEN
	}
	return localePTBR
}

// TranslateNotification renders a notification string for the given locale.
func TranslateNotification(locale, key string, params map[string]string) string {
	templates, ok := notificationTemplates[NormalizeNotificationLocale(locale)]
	if !ok {
		templates = notificationTemplates[localePTBR]
	}
	template, ok := templates[key]
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
func SlotLabel(locale string, journeyIndex int, isExit bool) string {
	number := fmt.Sprintf("%d", journeyIndex+1)
	if NormalizeNotificationLocale(locale) == localeEN {
		if isExit {
			return fmt.Sprintf("Exit %s", number)
		}
		return fmt.Sprintf("Entry %s", number)
	}
	if isExit {
		return fmt.Sprintf("Saída %s", number)
	}
	return fmt.Sprintf("Entrada %s", number)
}
