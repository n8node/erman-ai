import { fetchYandexMetrikaPublic } from "@/lib/yandex-metrika-server";
import { YandexMetrikaInjector } from "./YandexMetrikaInjector";

export async function YandexMetrika() {
  const data = await fetchYandexMetrikaPublic();
  if (!data?.enabled || !data.counter_code.trim()) return null;
  return <YandexMetrikaInjector counterCode={data.counter_code} />;
}
