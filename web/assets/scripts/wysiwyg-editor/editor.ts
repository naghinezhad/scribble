import { Editor } from "@tiptap/core";
import Underline from "@tiptap/extension-underline";
import BulletList from "@tiptap/extension-bullet-list";
import Text from "@tiptap/extension-text";
import Document from "@tiptap/extension-document";
import Blockquote from "@tiptap/extension-blockquote";
import CodeBlock from "@tiptap/extension-code-block";
import HardBreak from "@tiptap/extension-hard-break";
import Heading from "@tiptap/extension-heading";
import HorizontalRule from "@tiptap/extension-horizontal-rule";
import ListItem from "@tiptap/extension-list-item";
import OrderedList from "@tiptap/extension-ordered-list";
import Paragraph from "@tiptap/extension-paragraph";
import Bold from "@tiptap/extension-bold";
import Code from "@tiptap/extension-code";
import Italic from "@tiptap/extension-italic";
import Strike from "@tiptap/extension-strike";
import Dropcursor from "@tiptap/extension-dropcursor";
import Gapcursor from "@tiptap/extension-gapcursor";
import { Placeholder } from "@tiptap/extensions";
import { UndoRedo } from "@tiptap/extensions";
import { Markdown } from "@tiptap/markdown";
import Image from "@tiptap/extension-image";
import Link from "@tiptap/extension-link";

const initWysiwygEditor = (element: HTMLTextAreaElement) => {
    if (!element.id) {
        element.id = `textarea-${Math.random().toString(36).slice(2)}`;
    }

    const createDiv = (className: string, id?: string) => {
        const div = document.createElement("div");
        if (id) {
            div.id = id;
        }
        div.classList.add(className);
        return div;
    };

    const editorContainer = createDiv(
        "wysiwyg-editor",
        `${element.id}-wysiwyg-editor`
    );

    const editorContent = document.createElement("div");
    editorContent.id = `${element.id}-wysiwyg-editor-content`;

    editorContainer.append(editorContent);

    element.insertAdjacentElement("afterend", editorContainer);
    element.style.display = "none";

    const editor = new Editor({
        element: editorContent,
        extensions: [
            CodeBlock,
            Document,
            HardBreak,
            HorizontalRule,
            Text,
            Code,
            Markdown,
            Placeholder.configure({
                placeholder: "Write something...",
            }),

            Dropcursor,
            Gapcursor,
            UndoRedo,

            Heading.configure({
                levels: [2, 3],
            }),
            Paragraph,

            Bold,
            Italic,
            Underline,
            Strike,

            BulletList,
            OrderedList,
            ListItem,

            Blockquote,

            Link.configure({
                openOnClick: false,
            }),
            Image.configure({
                inline: true,
                allowBase64: true,
                resize: {
                    enabled: true,
                    alwaysPreserveAspectRatio: true,
                },
            }),
        ],
        content: element.value,
        contentType: "markdown",
        editorProps: {
            attributes: {
                class: "wysiwyg-editor-content",
            },
        },
        onUpdate({ editor }) {
            element.value = editor.getMarkdown();
        },
    });

    if (editor.isEmpty) {
        editor.commands.setContent("");
    }

    const elementLabel = document.querySelector(`label[for="${element.id}"]`);
    if (elementLabel) {
        if (!elementLabel.id) {
            elementLabel.id = `${element.id}-label`;
        }

        editorContent.setAttribute("aria-labelledby", elementLabel.id);

        elementLabel.addEventListener("click", () => {
            editor.commands.focus();
        });
    }
};

export { initWysiwygEditor };
