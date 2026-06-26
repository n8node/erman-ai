import Link from "next/link";
import { CalendarCheck, MessageCircle, ShieldCheck } from "lucide-react";

export default async function ConsultationConfirmedPage({
  params,
}: {
  params: Promise<{ metricId: string }>;
}) {
  const { metricId } = await params;

  return (
    <div className="mx-auto flex min-h-[calc(100vh-120px)] max-w-3xl items-center">
      <section className="w-full rounded-xl border border-border bg-bg p-6 shadow-sm">
        <div className="flex items-start gap-4">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-[#eaf3de] text-[#3b6d11]">
            <CalendarCheck className="h-5 w-5" />
          </div>
          <div>
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
              Консультация забронирована
            </p>
            <h1 className="mt-2 text-lg font-medium text-text">
              Оплата прошла успешно
            </h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-text2">
              Ваша консультация закреплена в расписании. Подтверждение и детали
              записи будут доступны в Telegram-боте.
            </p>
          </div>
        </div>

        <div className="mt-6 grid gap-3 md:grid-cols-2">
          <div className="rounded-lg border border-border bg-bg2 p-4">
            <div className="flex items-center gap-2 text-sm font-medium text-text">
              <MessageCircle className="h-4 w-4 text-accent" />
              Изменения по записи
            </div>
            <p className="mt-2 text-sm leading-6 text-text2">
              Если потребуется изменить тему, уточнить детали, перенести время
              или добавить вводные данные, отправьте информацию в Telegram-бота
              текстовым сообщением.
            </p>
          </div>

          <div className="rounded-lg border border-border bg-bg2 p-4">
            <div className="flex items-center gap-2 text-sm font-medium text-text">
              <ShieldCheck className="h-4 w-4 text-[#3b6d11]" />
              Что дальше
            </div>
            <p className="mt-2 text-sm leading-6 text-text2">
              Сохраните дату консультации. Если понадобится подготовить
              материалы заранее, пришлите их в бот в одном сообщении или
              несколькими короткими сообщениями.
            </p>
          </div>
        </div>

        <div className="mt-6 rounded-lg border border-border bg-bg2 px-4 py-3 text-xs text-text3">
          ID страницы для аналитики:{" "}
          <span className="font-mono text-text">{metricId}</span>
        </div>

        <div className="mt-6 flex flex-wrap gap-3">
          <Link
            href="/consultations"
            className="rounded-lg border border-border2 px-4 py-2 text-sm font-medium text-text hover:bg-bg2"
          >
            Вернуться к консультациям
          </Link>
          <Link
            href="/"
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90"
          >
            В дашборд
          </Link>
        </div>
      </section>
    </div>
  );
}
