"use client";

import LinkExtension from "@tiptap/extension-link";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import {
  Bold,
  Code2,
  Heading2,
  Italic,
  Link2,
  List,
  ListOrdered,
  Redo2,
  Undo2,
  Unlink,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

type Props = {
  value: string;
  onChange: (html: string) => void;
  editorKey: string;
};

type EditorMode = "visual" | "html";

function ToolbarButton({
  onClick,
  active,
  disabled,
  title,
  children,
}: {
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      title={title}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        "rounded-md p-1.5 text-text2 transition-colors hover:bg-bg2 hover:text-text disabled:opacity-40",
        active && "bg-bg2 text-text"
      )}
    >
      {children}
    </button>
  );
}

export function RichTextEditor({ value, onChange, editorKey }: Props) {
  const t = useTranslations("admin.publicPages.editor");
  const [mode, setMode] = useState<EditorMode>("visual");
  const [htmlSource, setHtmlSource] = useState(value);

  const editor = useEditor(
    {
      extensions: [
        StarterKit.configure({
          heading: { levels: [2, 3] },
        }),
        LinkExtension.configure({
          openOnClick: false,
          HTMLAttributes: { class: "text-accent underline" },
        }),
      ],
      content: value,
      immediatelyRender: false,
      onUpdate: ({ editor: ed }) => {
        onChange(ed.getHTML());
      },
      editorProps: {
        attributes: {
          class:
            "public-page-content min-h-[380px] px-4 py-3 outline-none focus:outline-none",
        },
      },
    },
    [editorKey]
  );

  useEffect(() => {
    if (!editor) return;
    const current = editor.getHTML();
    if (value !== current) {
      editor.commands.setContent(value || "<p></p>", { emitUpdate: false });
    }
    setHtmlSource(value);
  }, [editor, value, editorKey]);

  function switchToVisual() {
    if (editor) {
      editor.commands.setContent(htmlSource || "<p></p>", { emitUpdate: false });
      onChange(editor.getHTML());
    }
    setMode("visual");
  }

  function switchToHtml() {
    const html = editor?.getHTML() ?? value;
    setHtmlSource(html);
    onChange(html);
    setMode("html");
  }

  function setLink() {
    if (!editor) return;
    const previous = editor.getAttributes("link").href as string | undefined;
    const url = window.prompt(t("linkPrompt"), previous ?? "https://");
    if (url === null) return;
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      return;
    }
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
  }

  return (
    <div className="overflow-hidden rounded-lg border border-border2 bg-bg">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border bg-bg2 px-2 py-1.5">
        <div className="flex flex-wrap items-center gap-0.5">
          {mode === "visual" && editor && (
            <>
              <ToolbarButton
                title={t("bold")}
                active={editor.isActive("bold")}
                onClick={() => editor.chain().focus().toggleBold().run()}
              >
                <Bold className="h-4 w-4" />
              </ToolbarButton>
              <ToolbarButton
                title={t("italic")}
                active={editor.isActive("italic")}
                onClick={() => editor.chain().focus().toggleItalic().run()}
              >
                <Italic className="h-4 w-4" />
              </ToolbarButton>
              <span className="mx-1 h-4 w-px bg-border" />
              <ToolbarButton
                title={t("heading")}
                active={editor.isActive("heading", { level: 2 })}
                onClick={() =>
                  editor.chain().focus().toggleHeading({ level: 2 }).run()
                }
              >
                <Heading2 className="h-4 w-4" />
              </ToolbarButton>
              <ToolbarButton
                title={t("bulletList")}
                active={editor.isActive("bulletList")}
                onClick={() => editor.chain().focus().toggleBulletList().run()}
              >
                <List className="h-4 w-4" />
              </ToolbarButton>
              <ToolbarButton
                title={t("orderedList")}
                active={editor.isActive("orderedList")}
                onClick={() => editor.chain().focus().toggleOrderedList().run()}
              >
                <ListOrdered className="h-4 w-4" />
              </ToolbarButton>
              <span className="mx-1 h-4 w-px bg-border" />
              <ToolbarButton
                title={t("link")}
                active={editor.isActive("link")}
                onClick={setLink}
              >
                <Link2 className="h-4 w-4" />
              </ToolbarButton>
              <ToolbarButton
                title={t("unlink")}
                disabled={!editor.isActive("link")}
                onClick={() => editor.chain().focus().unsetLink().run()}
              >
                <Unlink className="h-4 w-4" />
              </ToolbarButton>
              <span className="mx-1 h-4 w-px bg-border" />
              <ToolbarButton
                title={t("undo")}
                disabled={!editor.can().chain().focus().undo().run()}
                onClick={() => editor.chain().focus().undo().run()}
              >
                <Undo2 className="h-4 w-4" />
              </ToolbarButton>
              <ToolbarButton
                title={t("redo")}
                disabled={!editor.can().chain().focus().redo().run()}
                onClick={() => editor.chain().focus().redo().run()}
              >
                <Redo2 className="h-4 w-4" />
              </ToolbarButton>
            </>
          )}
        </div>

        <div className="flex rounded-md border border-border bg-bg p-0.5 text-xs">
          <button
            type="button"
            onClick={switchToVisual}
            className={cn(
              "rounded px-2 py-1",
              mode === "visual" ? "bg-bg2 font-medium text-text" : "text-text2"
            )}
          >
            {t("visualMode")}
          </button>
          <button
            type="button"
            onClick={switchToHtml}
            className={cn(
              "inline-flex items-center gap-1 rounded px-2 py-1",
              mode === "html" ? "bg-bg2 font-medium text-text" : "text-text2"
            )}
          >
            <Code2 className="h-3 w-3" />
            {t("htmlMode")}
          </button>
        </div>
      </div>

      {mode === "visual" ? (
        <EditorContent editor={editor} />
      ) : (
        <textarea
          className="min-h-[420px] w-full resize-y border-0 bg-bg px-4 py-3 font-mono text-xs leading-relaxed text-text outline-none focus:ring-0"
          value={htmlSource}
          onChange={(e) => {
            setHtmlSource(e.target.value);
            onChange(e.target.value);
          }}
          spellCheck={false}
        />
      )}
    </div>
  );
}
