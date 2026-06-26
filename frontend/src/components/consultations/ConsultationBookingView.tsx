"use client";

import { useEffect, useMemo, useState } from "react";
import { CalendarDays, Clock, CreditCard } from "lucide-react";
import { useAuthUser } from "@/context/AuthContext";
import {
  createConsultationBooking,
  createConsultationCheckout,
  createPublicConsultationBooking,
  fetchConsultationSlots,
  fetchPublicConsultationServices,
  type ConsultationService,
  type ConsultationSlot,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function formatRub(value: number) {
  return new Intl.NumberFormat("ru-RU").format(value) + " ₽";
}

function formatDateLabel(value: Date) {
  return new Intl.DateTimeFormat("ru-RU", { weekday: "short", day: "2-digit", month: "short" }).format(value);
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}

function toDateInput(value: Date) {
  return value.toISOString().slice(0, 10);
}

export function ConsultationBookingView() {
  const user = useAuthUser();
  const [services, setServices] = useState<ConsultationService[]>([]);
  const [serviceId, setServiceId] = useState("");
  const [selectedDate, setSelectedDate] = useState(() => toDateInput(new Date()));
  const [slots, setSlots] = useState<ConsultationSlot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<ConsultationSlot | null>(null);
  const [name, setName] = useState("");
  const [email, setEmail] = useState(user?.email || "");
  const [phone, setPhone] = useState("");
  const [telegram, setTelegram] = useState("");
  const [note, setNote] = useState("");
  const [website, setWebsite] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchPublicConsultationServices()
      .then((data) => {
        setServices(data.items);
        setServiceId(data.items[0]?.id || "");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить услуги"))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (user?.email) setEmail(user.email);
  }, [user]);

  const service = services.find((item) => item.id === serviceId) || null;

  const dateOptions = useMemo(() => {
    const base = new Date();
    return Array.from({ length: 14 }, (_, i) => {
      const d = new Date(base);
      d.setDate(base.getDate() + i);
      return d;
    });
  }, []);

  useEffect(() => {
    if (!serviceId || !selectedDate) return;
    const from = new Date(`${selectedDate}T00:00:00`);
    const to = new Date(`${selectedDate}T23:59:59`);
    setSelectedSlot(null);
    fetchConsultationSlots(serviceId, from.toISOString(), to.toISOString())
      .then((data) => setSlots(data.items))
      .catch(() => setSlots([]));
  }, [serviceId, selectedDate]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!service || !selectedSlot) {
      setError("Выберите услугу, дату и время");
      return;
    }
    if (!name.trim() || !email.trim()) {
      setError("Укажите имя и email");
      return;
    }
    setSubmitting(true);
    try {
      const payload = {
        service_id: service.id,
        starts_at: selectedSlot.starts_at,
        customer_name: name.trim(),
        customer_email: email.trim(),
        customer_phone: phone.trim(),
        customer_telegram: telegram.trim(),
        customer_note: note.trim(),
        timezone: "Europe/Moscow",
        website,
      };
      const booking = user
        ? await createConsultationBooking(payload)
        : await createPublicConsultationBooking(payload);
      const checkout = await createConsultationCheckout(booking.id, !user);
      window.location.href = checkout.checkout_url;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось создать запись";
      setError(message === "slot unavailable" ? "Этот слот уже недоступен. Выберите другое время." : message);
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">Загрузка консультаций…</p>;
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div>
        <h1 className="text-base font-medium">Консультация</h1>
        <p className="mt-1 max-w-2xl text-sm text-text2">
          Выберите удобное время, оставьте контакты и оплатите консультацию. После оплаты запись закрепится, а подтверждение уйдёт администратору.
        </p>
      </div>

      {error && <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>}

      <form onSubmit={handleSubmit} className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-5">
          <section className="rounded-xl border border-border bg-bg p-5">
            <div className="mb-4 flex items-center gap-2">
              <CreditCard className="h-4 w-4 text-accent" />
              <h2 className="text-sm font-medium">Услуга</h2>
            </div>
            <div className="grid gap-3 md:grid-cols-2">
              {services.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => setServiceId(item.id)}
                  className={cn(
                    "rounded-xl border p-4 text-left transition hover:border-border2",
                    serviceId === item.id ? "border-text bg-bg2" : "border-border bg-bg"
                  )}
                >
                  <div className="flex items-start justify-between gap-3">
                    <h3 className="text-sm font-medium">{item.name}</h3>
                    <span className="rounded-full bg-bg2 px-2 py-1 text-xs text-text2">{item.duration_minutes} мин</span>
                  </div>
                  <p className="mt-2 line-clamp-3 text-sm text-text2">{item.description}</p>
                  <p className="mt-3 text-sm font-medium">{formatRub(item.price_rub)}</p>
                </button>
              ))}
            </div>
          </section>

          <section className="rounded-xl border border-border bg-bg p-5">
            <div className="mb-4 flex items-center gap-2">
              <CalendarDays className="h-4 w-4 text-accent" />
              <h2 className="text-sm font-medium">Дата</h2>
            </div>
            <div className="grid grid-cols-2 gap-2 md:grid-cols-4 lg:grid-cols-7">
              {dateOptions.map((date) => {
                const key = toDateInput(date);
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => setSelectedDate(key)}
                    className={cn(
                      "rounded-lg border px-3 py-2 text-sm",
                      selectedDate === key ? "border-text bg-text text-white" : "border-border bg-bg hover:bg-bg2"
                    )}
                  >
                    {formatDateLabel(date)}
                  </button>
                );
              })}
            </div>
          </section>

          <section className="rounded-xl border border-border bg-bg p-5">
            <div className="mb-4 flex items-center gap-2">
              <Clock className="h-4 w-4 text-accent" />
              <h2 className="text-sm font-medium">Время</h2>
            </div>
            {slots.filter((slot) => slot.available).length === 0 ? (
              <p className="text-sm text-text2">На выбранную дату нет свободных слотов.</p>
            ) : (
              <div className="grid grid-cols-2 gap-2 md:grid-cols-4">
                {slots.filter((slot) => slot.available).map((slot) => (
                  <button
                    key={slot.starts_at}
                    type="button"
                    onClick={() => setSelectedSlot(slot)}
                    className={cn(
                      "rounded-lg border px-3 py-2 text-sm",
                      selectedSlot?.starts_at === slot.starts_at
                        ? "border-accent bg-blue-50 text-accent"
                        : "border-border bg-bg hover:bg-bg2"
                    )}
                  >
                    {formatTime(slot.starts_at)}
                  </button>
                ))}
              </div>
            )}
          </section>
        </div>

        <aside className="space-y-5">
          <section className="rounded-xl border border-border bg-bg p-5">
            <h2 className="text-sm font-medium">Ваши данные</h2>
            <div className="mt-4 space-y-3">
              <input className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} placeholder="Имя" />
              <input className={fieldClass} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" type="email" />
              <input className={fieldClass} value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="Телефон" />
              <input className={fieldClass} value={telegram} onChange={(e) => setTelegram(e.target.value)} placeholder="Telegram / WhatsApp" />
              <textarea className={cn(fieldClass, "min-h-28")} value={note} onChange={(e) => setNote(e.target.value)} placeholder="Что хотите разобрать на консультации?" />
              <input className="hidden" value={website} onChange={(e) => setWebsite(e.target.value)} tabIndex={-1} autoComplete="off" />
            </div>
          </section>

          <section className="rounded-xl border border-border bg-bg p-5">
            <h2 className="text-sm font-medium">Итого</h2>
            <div className="mt-4 space-y-3 rounded-lg border border-border bg-bg2 p-4 text-sm">
              <div className="flex justify-between gap-4">
                <span className="text-text2">Услуга</span>
                <span className="text-right font-medium">{service?.name || "—"}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-text2">Время</span>
                <span className="text-right">
                  {selectedSlot ? `${new Date(selectedSlot.starts_at).toLocaleDateString("ru-RU")} · ${formatTime(selectedSlot.starts_at)}` : "—"}
                </span>
              </div>
              <div className="flex justify-between gap-4 border-t border-border pt-3">
                <span className="text-text2">К оплате</span>
                <span className="font-medium">{service ? formatRub(service.price_rub) : "—"}</span>
              </div>
            </div>
            <button
              type="submit"
              disabled={submitting || !service || !selectedSlot}
              className="mt-4 w-full rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
            >
              {submitting ? "Создаём запись…" : "Перейти к оплате"}
            </button>
            <p className="mt-3 text-xs text-text3">
              Слот резервируется на 15 минут. Запись подтверждается только после успешной оплаты.
            </p>
          </section>
        </aside>
      </form>
    </div>
  );
}
