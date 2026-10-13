window.oddsApp = function () {
    return {
        loading: false,
        timer: null,
        clockTimer: null,
        now: Date.now(),
        data: { games: [], stale: true, error: "" },

        init() {
            this.load();
            this.clockTimer = setInterval(() => (this.now = Date.now()), 1000);
            document.addEventListener("visibilitychange", () =>
                this.refreshWhenVisible(),
            );
            window.addEventListener("focus", () => this.refreshWhenVisible());
        },

        refreshWhenVisible() {
            if (document.hidden) return;

            clearTimeout(this.timer);
            this.load();
        },

        scheduleNextLoad() {
            const interval = Number(document.body.dataset.refreshIntervalMs);
            // Keep failed and stale snapshots on the same polling cadence. A
            // missing timestamp must not turn into a one-second retry loop.
            const delay = Math.max(interval, 1000);

            clearTimeout(this.timer);
            this.timer = setTimeout(() => this.load(), delay);
        },

        async load() {
            if (this.loading) return;

            this.loading = true;

            try {
                const response = await fetch("/api/odds", {
                    cache: "no-store",
                });

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
                this.scheduleNextLoad();
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

        hasSnapshot() {
            if (!this.data.fetchedAt) return false;

            const timestamp = new Date(this.data.fetchedAt).getTime();
            return Number.isFinite(timestamp) && timestamp > 0;
        },

        signed(value) {
            return value == null ? "" : (value > 0 ? "+" : "") + value;
        },

        previousPrice(price, market = "") {
            if (!price.previousOdds) return "";

            if (price.previousLine == null) {
                return `was ${price.previousOdds}`;
            }

            if (market === "total") {
                return `was ${this.totalLabel(price.label)} ${Math.abs(price.previousLine)} ${price.previousOdds}`;
            }

            return `was ${price.previousOdds} ${this.signed(price.previousLine)}`;
        },

        totalLabel(label) {
            return label === "Over" ? "O" : label === "Under" ? "U" : label;
        },

        isStale() {
            if (this.data.stale || !this.hasSnapshot()) return true;

            const interval = Number(document.body.dataset.refreshIntervalMs);
            const age = this.now - new Date(this.data.fetchedAt).getTime();
            return age > interval * 2;
        },

        snapshotAge() {
            if (!this.hasSnapshot()) return "";

            const age = Math.max(
                0,
                Math.floor(
                    (this.now - new Date(this.data.fetchedAt).getTime()) / 1000,
                ),
            );
            return `${age}s old`;
        },
    };
};
