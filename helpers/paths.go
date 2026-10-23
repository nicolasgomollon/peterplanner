//
//  peterplanner
//  Copyright (c) 2026 Nicolas Gomollon <nicolas@gomollon.me>
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package helpers

import (
	"os"
	"path/filepath"
)

func envPath(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return filepath.Clean(value)
	}
	return fallback
}

func ConfigPath() string {
	return envPath(
		"PETERPLANNER_CONFIG",
		"/etc/peterplanner/settings.json",
	)
}

func RegistrarDir() string {
	return envPath(
		"PETERPLANNER_REGISTRAR_DIR",
		"/data/registrar",
	)
}

func ReportsDir() string {
	return envPath(
		"PETERPLANNER_REPORTS_DIR",
		"/data/reports",
	)
}

func RegistrarPath(parts ...string) string {
	return filepath.Join(
		append([]string{RegistrarDir()}, parts...)...,
	)
}

func ReportPath(studentID string) string {
	return filepath.Join(
		ReportsDir(),
		"DGW_Report-"+studentID+".xsl",
	)
}
