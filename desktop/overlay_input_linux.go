//go:build linux

package main

/*
#cgo linux pkg-config: gtk+-3.0
#cgo CFLAGS: -w
#include <stdlib.h>
#include "gtk/gtk.h"

typedef struct OverlayInputArgs {
	int passthrough;
	int show;
} OverlayInputArgs;

static void apply_overlay_input(GtkWidget *widget, int passthrough) {
	if (!GTK_IS_WINDOW(widget)) {
		return;
	}
	gtk_window_set_accept_focus(GTK_WINDOW(widget), passthrough ? FALSE : TRUE);
	// GTK input shapes are best-effort across Linux compositors; if a window
	// manager ignores them, the webview-level pointer-events rule is the fallback.
	if (passthrough) {
		cairo_region_t *region = cairo_region_create();
		gtk_widget_input_shape_combine_region(widget, region);
		cairo_region_destroy(region);
	} else {
		gtk_widget_input_shape_combine_region(widget, NULL);
	}
}

static gboolean apply_overlay_input_args(gpointer data) {
	OverlayInputArgs *args = (OverlayInputArgs *)data;
	GList *windows = gtk_window_list_toplevels();
	for (GList *item = windows; item != NULL; item = item->next) {
		GtkWidget *widget = GTK_WIDGET(item->data);
		apply_overlay_input(widget, args->passthrough);
		if (args->show) {
			gtk_widget_show(widget);
			gtk_window_set_keep_above(GTK_WINDOW(widget), TRUE);
		}
	}
	g_list_free(windows);
	free(args);
	return G_SOURCE_REMOVE;
}

void PMSetOverlayInputPassthrough(int enabled) {
	OverlayInputArgs *args = (OverlayInputArgs *)malloc(sizeof(OverlayInputArgs));
	args->passthrough = enabled;
	args->show = 0;
	g_idle_add(apply_overlay_input_args, args);
}

int PMShowOverlayWindowPassive() {
	OverlayInputArgs *args = (OverlayInputArgs *)malloc(sizeof(OverlayInputArgs));
	args->passthrough = 1;
	args->show = 1;
	g_idle_add(apply_overlay_input_args, args);
	return 1;
}
*/
import "C"

func setOverlayInputPassthrough(enabled bool) {
	if enabled {
		C.PMSetOverlayInputPassthrough(1)
		return
	}
	C.PMSetOverlayInputPassthrough(0)
}

func showOverlayWindowPassive() bool {
	return C.PMShowOverlayWindowPassive() == 1
}
