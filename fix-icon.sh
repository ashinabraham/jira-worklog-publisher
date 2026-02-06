#!/bin/bash
# Post-build script to ensure the correct icon is embedded in the app bundle
# Wails seems to use a default iconfile.icns instead of appicon.icns

APP_BUNDLE="build/bin/JIRA Worklog Publisher.app"
ICON_SOURCE="appicon.icns"
ICON_TARGET="${APP_BUNDLE}/Contents/Resources/iconfile.icns"

if [ -f "$ICON_SOURCE" ] && [ -d "$APP_BUNDLE" ]; then
    echo "Copying $ICON_SOURCE to $ICON_TARGET..."
    cp "$ICON_SOURCE" "$ICON_TARGET"
    touch "$APP_BUNDLE"
    echo "Icon updated successfully!"
else
    echo "Error: Icon source or app bundle not found"
    exit 1
fi
