/** Anonymized client mark for case study (no real company identifiers). */
export function ClientBrandMark({ className = "" }: { className?: string }) {
  return (
    <span
      className={`inline-flex items-center gap-2.5 ${className}`}
      aria-label="Заказчик — нефтедобыча"
    >
      <svg
        width="32"
        height="32"
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className="shrink-0"
        aria-hidden
      >
        <rect width="32" height="32" rx="8" fill="#0E3A5C" />
        <path
          d="M16 6L22 14H18V24H14V14H10L16 6Z"
          fill="#4DA3D9"
          opacity="0.9"
        />
        <rect x="8" y="25" width="16" height="2" rx="1" fill="#7EC8F2" />
      </svg>
      <span className="text-[13px] font-medium tracking-[-0.01em] text-[#0B0D0E]">
        НефтеПром
      </span>
    </span>
  );
}
