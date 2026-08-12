package pkg

import (
	"strings"
	"testing"

	"github.com/spencer-osbrjp/bungkus-cli/config"
)

func setupRegistry(t *testing.T) {
	t.Helper()
	if err := InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
}

func TestFormLib_IsValidIntegration(t *testing.T) {
	setupRegistry(t)

	tests := []struct {
		name string
		form FormLib
		base string
		want bool
	}{
		{"none on any base", "none", "astro", true},
		{"react-hook-form on astro-react", "react-hook-form", "astro-react", true},
		{"react-hook-form on vite-react", "react-hook-form", "vite-react", true},
		{"react-hook-form on astro-vue", "react-hook-form", "astro-vue", false},
		{"react-hook-form on nuxt", "react-hook-form", "nuxt", false},
		{"react-hook-form on plain astro", "react-hook-form", "astro", false},
		{"veevalidate on astro-vue", "veevalidate", "astro-vue", true},
		{"veevalidate on nuxt (vue remap)", "veevalidate", "nuxt", true},
		{"veevalidate on astro-react", "veevalidate", "astro-react", false},
		{"tanstack-form on astro-react", "tanstack-form", "astro-react", true},
		{"tanstack-form on astro-vue", "tanstack-form", "astro-vue", true},
		{"tanstack-form on nuxt", "tanstack-form", "nuxt", true},
		{"tanstack-form on plain vite", "tanstack-form", "vite", false},
		{"unknown base", "react-hook-form", "bogus", false},
		{"unknown form", "bogus", "astro-react", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.form.IsValidIntegration(tt.base); got != tt.want {
				t.Errorf("IsValidIntegration(%q) = %v, want %v", tt.base, got, tt.want)
			}
		})
	}
}

func TestQueryLib_IsValidIntegration(t *testing.T) {
	setupRegistry(t)

	tests := []struct {
		name  string
		query QueryLib
		base  string
		want  bool
	}{
		{"none on any base", "none", "vite", true},
		{"tanstack-query on astro-react", "tanstack-query", "astro-react", true},
		{"tanstack-query on astro-vue", "tanstack-query", "astro-vue", true},
		{"tanstack-query on vite-react", "tanstack-query", "vite-react", true},
		{"tanstack-query on nuxt (vue remap)", "tanstack-query", "nuxt", true},
		{"tanstack-query on plain astro", "tanstack-query", "astro", false},
		{"tanstack-query on plain vite", "tanstack-query", "vite", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.IsValidIntegration(tt.base); got != tt.want {
				t.Errorf("IsValidIntegration(%q) = %v, want %v", tt.base, got, tt.want)
			}
		})
	}
}

func TestStateLib_IsValidIntegration(t *testing.T) {
	setupRegistry(t)

	tests := []struct {
		name  string
		state StateLib
		base  string
		want  bool
	}{
		{"none on any base", "none", "vite", true},
		{"jotai on vite-react", "jotai", "vite-react", true},
		{"jotai on astro-vue", "jotai", "astro-vue", false},
		{"jotai on nuxt", "jotai", "nuxt", false},
		{"zustand on astro-react", "zustand", "astro-react", true},
		{"zustand on nuxt", "zustand", "nuxt", false},
		{"pinia on astro-vue", "pinia", "astro-vue", true},
		{"pinia on nuxt (vue remap)", "pinia", "nuxt", true},
		{"pinia on vite-react", "pinia", "vite-react", false},
		{"nanostores on astro-react", "nanostores", "astro-react", true},
		{"nanostores on astro-vue", "nanostores", "astro-vue", true},
		{"nanostores on nuxt (vue remap)", "nanostores", "nuxt", true},
		{"nanostores on plain vite", "nanostores", "vite", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsValidIntegration(tt.base); got != tt.want {
				t.Errorf("IsValidIntegration(%q) = %v, want %v", tt.base, got, tt.want)
			}
		})
	}
}

func TestDesktopTarget_IsValid(t *testing.T) {
	setupRegistry(t)
	cases := map[DesktopTarget]bool{"none": true, "tauri": true, "electron": false, "": false}
	for d, want := range cases {
		if got := d.IsValid(); got != want {
			t.Errorf("DesktopTarget(%q).IsValid() = %v, want %v", d, got, want)
		}
	}
}

func TestCrateNameAndBundleIdentifier(t *testing.T) {
	cases := []struct {
		name  string
		crate string
		id    string
	}{
		{"my-app", "my-app", "com.example.my-app"},
		{"my.app", "my-app", "com.example.my-app"},
		{"My App", "my-app", "com.example.my-app"},
		{"my_app", "my-app", "com.example.my-app"},
		{"app", "app", "com.example.app-desktop"}, // .app suffix guard
		{"アプリ", "app", "com.example.app-desktop"},
	}
	for _, tc := range cases {
		c := ProjectConfig{ProjectName: tc.name}
		if got := c.CrateName(); got != tc.crate {
			t.Errorf("CrateName(%q) = %q, want %q", tc.name, got, tc.crate)
		}
		got := c.BundleIdentifier()
		if got != tc.id {
			t.Errorf("BundleIdentifier(%q) = %q, want %q", tc.name, got, tc.id)
		}
		if strings.HasSuffix(got, ".app") || got == "com.tauri.dev" {
			t.Errorf("BundleIdentifier(%q) = %q violates tauri identifier rules", tc.name, got)
		}
	}
}

func TestBaseIntegrationHelpers(t *testing.T) {
	setupRegistry(t)
	cases := []struct {
		base  BaseFramework
		react bool
		vue   bool
	}{
		{"astro", false, false},
		{"astro-react", true, false},
		{"astro-vue", false, true},
		{"vite", false, false},
		{"vite-react", true, false},
		{"vite-vue", false, true},
		// nuxt's registry entry has no integration field; its vue-ness is
		// remapped separately (see BuildPackageJSON / IsValidIntegration).
		{"nuxt", false, false},
		{"bogus", false, false},
	}
	for _, tc := range cases {
		if got := tc.base.IsReactInt(); got != tc.react {
			t.Errorf("%s.IsReactInt() = %v, want %v", tc.base, got, tc.react)
		}
		if got := tc.base.IsVueInt(); got != tc.vue {
			t.Errorf("%s.IsVueInt() = %v, want %v", tc.base, got, tc.vue)
		}
	}
}
