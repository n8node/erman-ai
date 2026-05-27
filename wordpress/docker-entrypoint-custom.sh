#!/bin/bash
set -euo pipefail

# Named volumes (and some bind mounts) are often root-owned on first use.
# WordPress runs PHP/Apache as www-data and must write uploads, themes, etc.
fix_wp_content_permissions() {
	local wp_content="/var/www/html/wp-content"
	for subdir in uploads themes plugins upgrade cache; do
		local dir="${wp_content}/${subdir}"
		mkdir -p "${dir}"
		chown -R www-data:www-data "${dir}"
		chmod -R ug+rwX "${dir}"
	done
}

fix_wp_content_permissions

exec docker-entrypoint.sh "$@"
