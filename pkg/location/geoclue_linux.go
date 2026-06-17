//go:build linux

package location

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	geoClueBusName       = "org.freedesktop.GeoClue2"
	geoClueManagerPath   = "/org/freedesktop/GeoClue2/Manager"
	geoClueManagerIface  = "org.freedesktop.GeoClue2.Manager"
	geoClueClientIface   = "org.freedesktop.GeoClue2.Client"
	geoClueLocationIface = "org.freedesktop.GeoClue2.Location"
	dbusPropertiesIface  = "org.freedesktop.DBus.Properties"
	desktopAppID         = "pm-planner"
)

func getDeviceLocation(ctx context.Context) (*DeviceLocation, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("%w: não foi possível acessar o serviço de localização do sistema", ErrUnavailable)
	}
	defer conn.Close()

	manager := conn.Object(geoClueBusName, dbus.ObjectPath(geoClueManagerPath))
	var clientPath dbus.ObjectPath
	if err := manager.Call(geoClueManagerIface+".GetClient", 0).Store(&clientPath); err != nil {
		return nil, fmt.Errorf("%w: GeoClue não está disponível (%v)", ErrUnavailable, err)
	}

	client := conn.Object(geoClueBusName, clientPath)
	if err := client.Call(
		dbusPropertiesIface+".Set", 0,
		geoClueClientIface,
		"DesktopId",
		dbus.MakeVariant(desktopAppID),
	).Err; err != nil {
		return nil, fmt.Errorf("%w: não foi possível configurar o GeoClue (%v)", ErrUnavailable, err)
	}
	if err := client.Call(geoClueClientIface+".Start", 0).Err; err != nil {
		return nil, mapGeoClueStartError(err)
	}
	defer func() {
		_ = client.Call(geoClueClientIface+".Stop", 0).Err
	}()

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var locationVariant dbus.Variant
		if err := client.Call(
			dbusPropertiesIface+".Get", 0,
			geoClueClientIface,
			"Location",
		).Store(&locationVariant); err != nil {
			return nil, fmt.Errorf("%w: falha ao consultar localização (%v)", ErrUnavailable, err)
		}

		locationPath, ok := locationVariant.Value().(dbus.ObjectPath)
		if ok && locationPath != "" && locationPath != "/" {
			loc, err := readGeoClueLocation(conn, locationPath)
			if err == nil {
				return loc, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}

	return nil, errors.New("tempo esgotado ao obter localização. Verifique se os serviços de localização do sistema estão habilitados (Configurações → Privacidade → Localização).")
}

func mapGeoClueStartError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "Geolocation disabled"):
		return errors.New("serviços de localização do sistema estão desativados. Habilite em Configurações → Privacidade → Serviços de localização.")
	case strings.Contains(msg, "Permission denied"), strings.Contains(msg, "Not authorized"):
		return errors.New("o app não tem permissão para acessar a localização do sistema. Verifique as permissões de privacidade.")
	default:
		return fmt.Errorf("%w: não foi possível iniciar o GeoClue (%v)", ErrUnavailable, err)
	}
}

func readGeoClueLocation(conn *dbus.Conn, path dbus.ObjectPath) (*DeviceLocation, error) {
	locObj := conn.Object(geoClueBusName, path)

	lat, err := readGeoClueDouble(locObj, "Latitude")
	if err != nil {
		return nil, err
	}
	lng, err := readGeoClueDouble(locObj, "Longitude")
	if err != nil {
		return nil, err
	}
	accuracy, err := readGeoClueDouble(locObj, "Accuracy")
	if err != nil {
		accuracy = 0
	}

	return &DeviceLocation{
		Latitude:  lat,
		Longitude: lng,
		Accuracy:  accuracy,
	}, nil
}

func readGeoClueDouble(locObj dbus.BusObject, property string) (float64, error) {
	var valueVariant dbus.Variant
	if err := locObj.Call(
		dbusPropertiesIface+".Get", 0,
		geoClueLocationIface,
		property,
	).Store(&valueVariant); err != nil {
		return 0, err
	}
	switch v := valueVariant.Value().(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("tipo inesperado para %s", property)
	}
}
