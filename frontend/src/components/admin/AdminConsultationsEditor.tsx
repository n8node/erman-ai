"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  createAdminConsultationService,
  fetchAdminConsultationAvailability,
  fetchAdminConsultationBookings,
  fetchAdminConsultationServices,
  updateAdminConsultationAvailability,
  updateAdminConsultationBookingStatus,
  updateAdminConsultationService,
  type ConsultationAvailabilityRule,
  type ConsultationBooking,
  type ConsultationBookingDetail,
  type ConsultationService,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const emptyService: Partial<ConsultationService> = {
  slug: "",
  name: "",
  description: "",
  duration_minutes: 90,
  price_rub: 100,
  is_active: true,
  min_notice_minutes: 180,
  max_advance_days: 30,
  buffer_before_minutes: 0,
  buffer_after_minutes: 0,
  meeting_url: "",
  sort_order: 100,
};

const weekdays = [
  ["0", "Вс"],
  ["1", "Пн"],
  ["2", "Вт"],
  ["3", "Ср"],
  ["4", "Чт"],
  ["5", "Пт"],
  ["6", "Сб"],
];

function formatDate(value: string) {
  return new Intl.DateTimeFormat("ru-RU", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}

function formatRub(value: number) {
  return new Intl.NumberFormat("ru-RU").format(value) + " ₽";
}

export function AdminConsultationsEditor() {
  const [services, setServices] = useState<ConsultationService[]>([]);
  const [bookings, setBookings] = useState<ConsultationBookingDetail[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [form, setForm] = useState<Partial<ConsultationService>>(emptyService);
  const [availability, setAvailability] = useState<ConsultationAvailabilityRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const selected = useMemo(
    () => services.find((item) => item.id === selectedId) || null,
    [services, selectedId]
  );

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [svc, bookingData] = await Promise.all([
        fetchAdminConsultationServices(),
        fetchAdminConsultationBookings({ limit: 50, offset: 0 }),
      ]);
      setServices(svc.items);
      setBookings(bookingData.items);
      const first = svc.items[0];
      if (first) {
        setSelectedId((prev) => prev || first.id);
        setForm(first);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить консультации");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!selectedId) {
      setAvailability([]);
      return;
    }
    fetchAdminConsultationAvailability(selectedId)
      .then((data) => setAvailability(data.items))
      .catch(() => setAvailability([]));
  }, [selectedId]);

  function selectService(item: ConsultationService) {
    setSelectedId(item.id);
    setForm(item);
    setSuccess("");
    setError("");
  }

  function newService() {
    setSelectedId("");
    setForm(emptyService);
    setAvailability([]);
    setSuccess("");
    setError("");
  }

  async function saveService() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const payload = {
        ...form,
        duration_minutes: Number(form.duration_minutes || 0),
        price_rub: Number(form.price_rub || 0),
        min_notice_minutes: Number(form.min_notice_minutes || 0),
        max_advance_days: Number(form.max_advance_days || 30),
        buffer_before_minutes: Number(form.buffer_before_minutes || 0),
        buffer_after_minutes: Number(form.buffer_after_minutes || 0),
        sort_order: Number(form.sort_order || 100),
      };
      const saved = selectedId
        ? await updateAdminConsultationService(selectedId, payload)
        : await createAdminConsultationService(payload);
      setServices((prev) => {
        const exists = prev.some((item) => item.id === saved.id);
        return exists ? prev.map((item) => (item.id === saved.id ? saved : item)) : [saved, ...prev];
      });
      setSelectedId(saved.id);
      setForm(saved);
      setSuccess("Услуга сохранена");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить услугу");
    } finally {
      setSaving(false);
    }
  }

  async function saveAvailability() {
    if (!selectedId) {
      setError("Сначала сохраните услугу");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const saved = await updateAdminConsultationAvailability(selectedId, availability);
      setAvailability(saved.items);
      setSuccess("Расписание сохранено");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить расписание");
    } finally {
      setSaving(false);
    }
  }

  function updateRule(index: number, patch: Partial<ConsultationAvailabilityRule>) {
    setAvailability((prev) => prev.map((item, i) => (i === index ? { ...item, ...patch } : item)));
  }

  function addRule() {
    setAvailability((prev) => [
      ...prev,
      { weekday: 1, start_time: "10:00", end_time: "18:00", slot_step_minutes: 15, is_active: true },
    ]);
  }

  async function changeBookingStatus(id: string, status: ConsultationBooking["status"]) {
    try {
      const updated = await updateAdminConsultationBookingStatus(id, status);
      setBookings((prev) => prev.map((item) => (item.id === id ? updated : item)));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось обновить бронь");
    }
  }

  if (loading) return <p className="text-sm text-text2">Загрузка…</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-base font-medium">Консультации</h1>
          <p className="mt-1 text-sm text-text2">Услуги, стоимость, рабочее расписание и оплаченные записи.</p>
        </div>
        <button type="button" onClick={newService} className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white">
          Новая услуга
        </button>
      </div>

      {error && <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>}
      {success && <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">{success}</div>}

      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        <aside className="rounded-xl border border-border bg-bg p-4">
          <h2 className="text-sm font-medium">Услуги</h2>
          <div className="mt-3 space-y-2">
            {services.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => selectService(item)}
                className={cn(
                  "w-full rounded-lg border px-3 py-2 text-left text-sm",
                  selectedId === item.id ? "border-text bg-bg2" : "border-border hover:bg-bg2"
                )}
              >
                <span className="block font-medium">{item.name}</span>
                <span className="text-xs text-text2">{item.duration_minutes} мин · {formatRub(item.price_rub)}</span>
              </button>
            ))}
          </div>
        </aside>

        <main className="space-y-6">
          <section className="rounded-xl border border-border bg-bg p-5">
            <h2 className="text-sm font-medium">{selected ? "Редактировать услугу" : "Новая услуга"}</h2>
            <div className="mt-4 grid gap-4 md:grid-cols-2">
              <label className="space-y-1 text-sm">
                <span className="text-text2">Название</span>
                <input className={fieldClass} value={form.name || ""} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} />
              </label>
              <label className="space-y-1 text-sm">
                <span className="text-text2">Slug</span>
                <input className={fieldClass} value={form.slug || ""} onChange={(e) => setForm((p) => ({ ...p, slug: e.target.value }))} />
              </label>
              <label className="space-y-1 text-sm md:col-span-2">
                <span className="text-text2">Описание</span>
                <textarea className={cn(fieldClass, "min-h-24")} value={form.description || ""} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} />
              </label>
              <label className="space-y-1 text-sm">
                <span className="text-text2">Длительность, минут</span>
                <input className={fieldClass} type="number" min={1} value={form.duration_minutes || 0} onChange={(e) => setForm((p) => ({ ...p, duration_minutes: Number(e.target.value) }))} />
                <div className="flex gap-2 pt-1">
                  {[30, 60, 90].map((value) => (
                    <button key={value} type="button" onClick={() => setForm((p) => ({ ...p, duration_minutes: value }))} className="rounded border border-border2 px-2 py-1 text-xs hover:bg-bg2">
                      {value} мин
                    </button>
                  ))}
                </div>
              </label>
              <label className="space-y-1 text-sm">
                <span className="text-text2">Цена, ₽</span>
                <input className={fieldClass} type="number" min={0} value={form.price_rub || 0} onChange={(e) => setForm((p) => ({ ...p, price_rub: Number(e.target.value) }))} />
              </label>
              <label className="space-y-1 text-sm">
                <span className="text-text2">Минимум до записи, минут</span>
                <input className={fieldClass} type="number" min={0} value={form.min_notice_minutes || 0} onChange={(e) => setForm((p) => ({ ...p, min_notice_minutes: Number(e.target.value) }))} />
              </label>
              <label className="space-y-1 text-sm">
                <span className="text-text2">Горизонт записи, дней</span>
                <input className={fieldClass} type="number" min={1} value={form.max_advance_days || 30} onChange={(e) => setForm((p) => ({ ...p, max_advance_days: Number(e.target.value) }))} />
              </label>
              <label className="space-y-1 text-sm md:col-span-2">
                <span className="text-text2">Ссылка на встречу (опционально)</span>
                <input className={fieldClass} value={form.meeting_url || ""} onChange={(e) => setForm((p) => ({ ...p, meeting_url: e.target.value }))} />
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={!!form.is_active} onChange={(e) => setForm((p) => ({ ...p, is_active: e.target.checked }))} />
                Услуга активна
              </label>
            </div>
            <button type="button" disabled={saving} onClick={saveService} className="mt-4 rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-60">
              Сохранить услугу
            </button>
          </section>

          <section className="rounded-xl border border-border bg-bg p-5">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-medium">Рабочее расписание</h2>
              <button type="button" onClick={addRule} className="rounded border border-border2 px-3 py-1 text-sm hover:bg-bg2">Добавить интервал</button>
            </div>
            <div className="mt-4 space-y-2">
              {availability.map((rule, index) => (
                <div key={`${rule.weekday}-${index}`} className="grid gap-2 rounded-lg border border-border bg-bg2 p-3 md:grid-cols-[90px_1fr_1fr_1fr_80px]">
                  <select className={fieldClass} value={rule.weekday} onChange={(e) => updateRule(index, { weekday: Number(e.target.value) })}>
                    {weekdays.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                  </select>
                  <input className={fieldClass} type="time" value={(rule.start_time || "").slice(0, 5)} onChange={(e) => updateRule(index, { start_time: e.target.value })} />
                  <input className={fieldClass} type="time" value={(rule.end_time || "").slice(0, 5)} onChange={(e) => updateRule(index, { end_time: e.target.value })} />
                  <input className={fieldClass} type="number" min={1} value={rule.slot_step_minutes} onChange={(e) => updateRule(index, { slot_step_minutes: Number(e.target.value) })} />
                  <button type="button" onClick={() => setAvailability((prev) => prev.filter((_, i) => i !== index))} className="rounded border border-border2 text-sm hover:bg-bg">Удалить</button>
                </div>
              ))}
            </div>
            <button type="button" disabled={saving || !selectedId} onClick={saveAvailability} className="mt-4 rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-60">
              Сохранить расписание
            </button>
          </section>
        </main>
      </div>

      <section className="rounded-xl border border-border bg-bg p-5">
        <h2 className="text-sm font-medium">Записи</h2>
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-3 py-2 font-medium">Дата</th>
                <th className="px-3 py-2 font-medium">Услуга</th>
                <th className="px-3 py-2 font-medium">Клиент</th>
                <th className="px-3 py-2 font-medium">Контакты</th>
                <th className="px-3 py-2 font-medium">Сумма</th>
                <th className="px-3 py-2 font-medium">Статус</th>
              </tr>
            </thead>
            <tbody>
              {bookings.map((booking) => (
                <tr key={booking.id} className="border-b border-border last:border-0">
                  <td className="px-3 py-3 text-text2">{formatDate(booking.starts_at)}</td>
                  <td className="px-3 py-3">{booking.service_name}</td>
                  <td className="px-3 py-3">{booking.customer_name}</td>
                  <td className="px-3 py-3 text-text2">
                    <div>{booking.customer_email}</div>
                    <div>{booking.customer_phone || booking.customer_telegram || "—"}</div>
                  </td>
                  <td className="px-3 py-3">{formatRub(booking.amount_rub)}</td>
                  <td className="px-3 py-3">
                    <select className="rounded border border-border2 px-2 py-1 text-xs" value={booking.status} onChange={(e) => changeBookingStatus(booking.id, e.target.value as ConsultationBooking["status"])}>
                      <option value="pending_payment">Ожидает оплату</option>
                      <option value="paid">Оплачена</option>
                      <option value="cancelled">Отменена</option>
                      <option value="expired">Истекла</option>
                      <option value="completed">Завершена</option>
                      <option value="no_show">Не пришёл</option>
                    </select>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {bookings.length === 0 && <p className="py-6 text-sm text-text2">Пока нет записей.</p>}
        </div>
      </section>
    </div>
  );
}
