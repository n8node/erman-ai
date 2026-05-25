/**
 * Adds missing i18n keys for calculator budget lines, recommendations, payment schedules.
 * Usage: node scripts/patch-locale-keys.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.join(__dirname, "../frontend/src/messages");

const patches = {
  en: {
    calculator: {
      recommendation: {
        automate: "Automation is recommended — payback under 12 months and positive ROI",
        consider: "Consider automation — moderate payback or ROI",
        not_recommended: "Automation is not recommended at current parameters",
      },
      budgetEstimate: {
        lines: {
          audit: "Discovery and project management",
          dev: "Development (~{hours} h × {rate} ₽)",
          integrations: "Integrations and systems",
          training: "Training and go-live",
          factorComplex: "Complex integrations (+{pct}%)",
          factorSimple: "Simple scope (−{pct}%)",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "50% prepayment, 50% on completion",
          "100_prepaid": "100% prepayment",
          "30_70": "30% prepayment, 70% on completion",
          milestone: "Milestone-based payments",
        },
        calcPrefill: {
          solutionName: "Automation: {process}",
          clientProblem:
            "Process «{process}»: potential savings {savings} ₽/mo after automation.",
          processFallback: "process",
        },
      },
    },
  },
  ru: {
    calculator: {
      recommendation: {
        automate: "Рекомендуется автоматизировать — окупаемость менее 12 месяцев и положительный ROI",
        consider: "Рассмотрите автоматизацию — умеренная окупаемость или ROI",
        not_recommended: "Автоматизация не рекомендуется при текущих параметрах",
      },
      budgetEstimate: {
        lines: {
          audit: "Аудит и управление проектом",
          dev: "Разработка (~{hours} ч × {rate} ₽)",
          integrations: "Интеграции и контуры",
          training: "Обучение и запуск",
          factorComplex: "Сложные интеграции (+{pct}%)",
          factorSimple: "Простой контур (−{pct}%)",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "50% предоплата, 50% по завершению",
          "100_prepaid": "100% предоплата",
          "30_70": "30% предоплата, 70% по завершению",
          milestone: "Поэтапная оплата по milestone",
        },
        calcPrefill: {
          solutionName: "Автоматизация: {process}",
          clientProblem:
            "Процесс «{process}»: потенциальная экономия {savings} ₽/мес после автоматизации.",
          processFallback: "процесс",
        },
      },
    },
  },
  de: {
    calculator: {
      recommendation: {
        automate: "Automatisierung empfohlen — Amortisation unter 12 Monaten und positiver ROI",
        consider: "Automatisierung erwägen — moderate Amortisation oder ROI",
        not_recommended: "Automatisierung bei den aktuellen Parametern nicht empfohlen",
      },
      budgetEstimate: {
        lines: {
          audit: "Analyse und Projektmanagement",
          dev: "Entwicklung (~{hours} Std. × {rate} ₽)",
          integrations: "Integrationen und Systeme",
          training: "Schulung und Go-Live",
          factorComplex: "Komplexe Integrationen (+{pct}%)",
          factorSimple: "Einfacher Umfang (−{pct}%)",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "50 % Vorauszahlung, 50 % bei Abschluss",
          "100_prepaid": "100 % Vorauszahlung",
          "30_70": "30 % Vorauszahlung, 70 % bei Abschluss",
          milestone: "Zahlung nach Meilensteinen",
        },
        calcPrefill: {
          solutionName: "Automatisierung: {process}",
          clientProblem:
            "Prozess «{process}»: potenzielle Einsparung {savings} ₽/Monat nach der Automatisierung.",
          processFallback: "Prozess",
        },
      },
    },
  },
  es: {
    calculator: {
      recommendation: {
        automate: "Se recomienda la automatización — recuperación en menos de 12 meses y ROI positivo",
        consider: "Considere la automatización — recuperación o ROI moderados",
        not_recommended: "La automatización no se recomienda con los parámetros actuales",
      },
      budgetEstimate: {
        lines: {
          audit: "Análisis y gestión del proyecto",
          dev: "Desarrollo (~{hours} h × {rate} ₽)",
          integrations: "Integraciones y sistemas",
          training: "Formación y puesta en marcha",
          factorComplex: "Integraciones complejas (+{pct}%)",
          factorSimple: "Alcance simple (−{pct}%)",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "50 % prepago, 50 % al finalizar",
          "100_prepaid": "100 % prepago",
          "30_70": "30 % prepago, 70 % al finalizar",
          milestone: "Pago por hitos",
        },
        calcPrefill: {
          solutionName: "Automatización: {process}",
          clientProblem:
            "Proceso «{process}»: ahorro potencial de {savings} ₽/mes tras la automatización.",
          processFallback: "proceso",
        },
      },
    },
  },
  fr: {
    calculator: {
      recommendation: {
        automate: "Automatisation recommandée — retour sur investissement en moins de 12 mois et ROI positif",
        consider: "Envisagez l'automatisation — retour sur investissement ou ROI modérés",
        not_recommended: "L'automatisation n'est pas recommandée avec les paramètres actuels",
      },
      budgetEstimate: {
        lines: {
          audit: "Audit et gestion de projet",
          dev: "Développement (~{hours} h × {rate} ₽)",
          integrations: "Intégrations et systèmes",
          training: "Formation et mise en production",
          factorComplex: "Intégrations complexes (+{pct}%)",
          factorSimple: "Périmètre simple (−{pct}%)",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "50 % d'acompte, 50 % à la fin",
          "100_prepaid": "100 % d'acompte",
          "30_70": "30 % d'acompte, 70 % à la fin",
          milestone: "Paiement par jalons",
        },
        calcPrefill: {
          solutionName: "Automatisation : {process}",
          clientProblem:
            "Processus « {process} » : économies potentielles de {savings} ₽/mois après automatisation.",
          processFallback: "processus",
        },
      },
    },
  },
  zh: {
    calculator: {
      recommendation: {
        automate: "建议自动化——回本周期少于12个月且ROI为正",
        consider: "可考虑自动化——回本周期或ROI适中",
        not_recommended: "当前参数下不建议自动化",
      },
      budgetEstimate: {
        lines: {
          audit: "调研与项目管理",
          dev: "开发（约 {hours} 小时 × {rate} ₽）",
          integrations: "集成与系统",
          training: "培训与上线",
          factorComplex: "复杂集成（+{pct}%）",
          factorSimple: "简单范围（−{pct}%）",
        },
      },
    },
    proposal: {
      form: {
        paymentSchedules: {
          "50_50": "预付50%，完成时付50%",
          "100_prepaid": "100% 预付",
          "30_70": "预付30%，完成时付70%",
          milestone: "按里程碑付款",
        },
        calcPrefill: {
          solutionName: "自动化：{process}",
          clientProblem: "流程「{process}」：自动化后潜在节省 {savings} ₽/月。",
          processFallback: "流程",
        },
      },
    },
  },
};

function deepMerge(target, source) {
  for (const [key, value] of Object.entries(source)) {
    if (value && typeof value === "object" && !Array.isArray(value)) {
      if (!target[key] || typeof target[key] !== "object") target[key] = {};
      deepMerge(target[key], value);
    } else {
      target[key] = value;
    }
  }
}

for (const [locale, patch] of Object.entries(patches)) {
  const filePath = path.join(messagesDir, `${locale}.json`);
  const data = JSON.parse(fs.readFileSync(filePath, "utf8"));
  deepMerge(data, patch);
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + "\n", "utf8");
  console.log(`Patched ${filePath}`);
}

console.log("Done.");
