const HOME_URL = "https://erman.ai";
const LOGO_SRC = "/dashboard/icon.png";

type LandingBrandLinkProps = {
  className?: string;
  labelClassName?: string;
};

export function LandingBrandLink({
  className = "",
  labelClassName = "",
}: LandingBrandLinkProps) {
  return (
    <a
      href={HOME_URL}
      aria-label="Erman AI"
      className={["inline-flex items-center gap-2.5 font-semibold", className].filter(Boolean).join(" ")}
    >
      <span className="grid h-[30px] w-[30px] shrink-0 place-items-center">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={LOGO_SRC} alt="" className="max-h-[30px] max-w-[30px] object-contain" />
      </span>
      <span className={labelClassName}>Erman AI</span>
    </a>
  );
}
