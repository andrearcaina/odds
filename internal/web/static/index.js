window.oddsApp = function () {
    return {
        loading: false,
        timer: null,
        data: { games: [], stale: true, error: "" },

        init() {
            this.load();
            const interval = Number(document.body.dataset.refreshIntervalMs);
            this.timer = setInterval(() => this.load(), interval);
        },

        async load(manualRefresh = false) {
            if (this.loading) return;

            this.loading = true;

            try {
                const response = await fetch(
                    "/api/odds" + (manualRefresh ? "?refresh=1" : ""),
                    { cache: "no-store" },
                );

                if (!response.ok) {
                    throw new Error("App API returned HTTP " + response.status);
                }

                this.data = await response.json();
            } catch (error) {
                this.data = {
                    ...this.data,
                    stale: true,
                    error: "Unable to reach the app API",
                };
            } finally {
                this.loading = false;
            }
        },

        formatTime(value) {
            return value
                ? new Date(value).toLocaleString("en-CA", {
                      timeZone: "America/Toronto",
                      weekday: "short",
                      month: "short",
                      day: "numeric",
                      hour: "numeric",
                      minute: "2-digit",
                      second: "2-digit",
                      timeZoneName: "short",
                  })
                : "";
        },

        signed(value) {
            return value == null ? "" : (value > 0 ? "+" : "") + value;
        },
    };
};
