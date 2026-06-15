export type YandexMetrikaPublic = {
  enabled: boolean;
  counter_code: string;
};

function serverApiBase() {
  return (
    process.env.INTERNAL_API_URL?.replace(/\/$/, "") || "http://backend:8080"
  );
}

export async function fetchYandexMetrikaPublic(): Promise<YandexMetrikaPublic | null> {
  try {
    const res = await fetch(
      `${serverApiBase()}/api/v1/public/yandex-metrika`,
      { next: { revalidate: 60 } }
    );
    if (!res.ok) return null;
    return (await res.json()) as YandexMetrikaPublic;
  } catch {
    return null;
  }
}
