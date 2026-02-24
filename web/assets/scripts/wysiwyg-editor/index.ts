import { initWysiwygEditor } from "./editor";

const selector = "textarea[data-wysiwyg-editor]";

const initEditors = (node: Node) => {
    if (node instanceof HTMLTextAreaElement && node.matches(selector)) {
        initWysiwygEditor(node);
        return;
    }

    if (
        node instanceof Element ||
        node instanceof Document ||
        node instanceof DocumentFragment
    ) {
        node.querySelectorAll<HTMLTextAreaElement>(selector).forEach(
            initWysiwygEditor
        );
    }
};

const observer = new MutationObserver((mutations) => {
    mutations.forEach(({ addedNodes }) => {
        addedNodes.forEach((addedNode) => initEditors(addedNode));
    });
});

if (document.body) {
    observer.observe(document.body, { childList: true, subtree: true });
}

initEditors(document);
