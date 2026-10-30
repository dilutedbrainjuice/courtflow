"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
type Court = { id: number; name: string; price_sen: number };
type Slot = {
  start_at: string;
  end_at: string;
  available: boolean;
  price_sen: number;
};
const base = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
const money = (sen: number) =>
  new Intl.NumberFormat("en-MY", { style: "currency", currency: "MYR" }).format(
    sen / 100,
  );
const localDate = (date: Date) =>
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Kuala_Lumpur",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
const hour = (value: string) =>
  new Intl.DateTimeFormat("en-MY", {
    timeZone: "Asia/Kuala_Lumpur",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
async function request(path: string, options?: RequestInit) {
  const response = await fetch(base + path, {
    ...options,
    signal: options?.signal
      ? AbortSignal.any([options.signal, AbortSignal.timeout(15000)])
      : AbortSignal.timeout(15000),
  });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || "Request failed");
  return data;
}
export default function Home() {
  const [today] = useState(() => localDate(new Date()));
  const [lastDay] = useState(() =>
    localDate(new Date(Date.now() + 30 * 86400000)),
  );
  const [courts, setCourts] = useState<Court[]>([]),
    [court, setCourt] = useState(1),
    [date, setDate] = useState(() => localDate(new Date())),
    [slots, setSlots] = useState<Slot[]>([]),
    [selected, setSelected] = useState<Slot | null>(null),
    [loading, setLoading] = useState(true),
    [pending, setPending] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [refresh, setRefresh] = useState(0),
    [venue, setVenue] = useState("CourtFlow Demo Badminton");
  useEffect(() => {
    const ctrl = new AbortController();
    request("/api/v1/facilities", { signal: ctrl.signal })
      .then((data) => {
        if (data[0]) setVenue(data[0].name);
        return request("/api/v1/facilities/1/courts", { signal: ctrl.signal });
      })
      .then(setCourts)
      .catch((e) => {
        if (!ctrl.signal.aborted) setError(e.message);
      });
    return () => ctrl.abort();
  }, [refresh]);
  useEffect(() => {
    const ctrl = new AbortController();
    request(`/api/v1/courts/${court}/availability?date=${date}`, {
      signal: ctrl.signal,
    })
      .then(setSlots)
      .catch((e) => {
        if (!ctrl.signal.aborted) setError(e.message);
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setLoading(false);
      });
    return () => ctrl.abort();
  }, [court, date, refresh]);
  function reload() {
    setLoading(true);
    setSelected(null);
    setSlots([]);
    setError("");
    setRefresh((v) => v + 1);
  }
  async function book() {
    if (!selected) return;
    setPending(true);
    setError("");
    setNotice("");
    try {
      const result = await request("/api/v1/bookings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ court_id: court, start_at: selected.start_at }),
      });
      setNotice(
        `Demo confirmed · ${money(result.amount_sen)} · Reference ${result.id}`,
      );
      reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Network unavailable");
      setSelected(null);
      setSlots([]);
      setLoading(true);
      setRefresh((v) => v + 1);
    } finally {
      setPending(false);
    }
  }
  return (
    <main>
      <header>
        <Link className="brand" href="/">
          ◈ CourtFlow
        </Link>
        <span className="badge">DEMO · NO REAL PAYMENTS</span>
      </header>
      <section className="hero">
        <p className="eyebrow">LESS PLANNING. MORE PLAYING.</p>
        <h1>
          Your court.
          <br />
          <em>Your next game.</em>
        </h1>
        <p>
          Book a little time for a great game. Four badminton courts, one simple
          way to play.
        </p>
        <div className="venue">
          ↗ {venue}
          <small>Malaysia · All times in Kuala Lumpur (MYT)</small>
        </div>
      </section>
      <section className="booking" aria-label="Book a court">
        <div className="sectionTitle">
          <div>
            <p className="eyebrow">LET’S GET YOU ON COURT</p>
            <h2>Choose your session</h2>
          </div>
          <button className="secondary" onClick={reload} disabled={pending}>
            Refresh
          </button>
        </div>
        <div className="controls">
          <label>
            Playing date
            <input
              type="date"
              value={date}
              min={today}
              max={lastDay}
              disabled={pending}
              onChange={(e) => {
                setDate(e.target.value);
                setSelected(null);
                setSlots([]);
                setLoading(true);
                setError("");
                setNotice("");
              }}
            />
          </label>
          <fieldset disabled={pending}>
            <legend>Your court</legend>
            <div className="courts">
              {courts.map((c) => (
                <button
                  className={court === c.id ? "active" : ""}
                  aria-pressed={court === c.id}
                  key={c.id}
                  onClick={() => {
                    setCourt(c.id);
                    setSelected(null);
                    setSlots([]);
                    setLoading(true);
                    setError("");
                    setNotice("");
                  }}
                >
                  {c.name}
                  <small>{money(c.price_sen)} / hour</small>
                </button>
              ))}
            </div>
          </fieldset>
        </div>
        <div aria-live="polite">
          {error && (
            <p className="error" role="alert">
              {error}. Refresh to try again.
            </p>
          )}
          {notice && (
            <p className="success" role="status">
              {notice}
            </p>
          )}
        </div>
        <h3>
          Available starts <span>1-hour sessions</span>
        </h3>
        {loading ? (
          <p role="status">Finding your next game…</p>
        ) : (
          <div className="slots">
            {slots.map((s) => (
              <button
                key={s.start_at}
                disabled={!s.available || pending}
                className={selected?.start_at === s.start_at ? "active" : ""}
                aria-pressed={selected?.start_at === s.start_at}
                onClick={() => setSelected(s)}
              >
                {hour(s.start_at)}
                <small>{s.available ? "Available" : "Unavailable"}</small>
              </button>
            ))}
          </div>
        )}
        <div className="review">
          <div>
            <strong>
              {selected
                ? `${hour(selected.start_at)} – ${hour(selected.end_at)} · ${money(selected.price_sen)}`
                : "Pick a time. Bring your game."}
            </strong>
            <p>Guest demo booking. No personal details or payment needed.</p>
          </div>
          <button
            className="primary"
            disabled={!selected || pending || loading}
            onClick={book}
          >
            {pending ? "Confirming…" : "Confirm demo booking →"}
          </button>
        </div>
      </section>
      <footer>
        Built for the love of the game.{" "}
        <span>Fictional venue · Public portfolio demo</span>
      </footer>
    </main>
  );
}
