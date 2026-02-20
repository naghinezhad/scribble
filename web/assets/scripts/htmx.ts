import htmx from "htmx.org";

window.htmx = htmx;

function showErrorMessage(text: string) {
    const div = document.createElement("div");
    div.className = "as-message type-error";
    div.setAttribute("role", "alert");
    div.innerHTML = `<div><svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2L1 21h22M12 6l7.53 13H4.47M11 10v4h2v-4m-2 6v2h2v-2"/></svg></div><div>${text}</div><div></div>`;
    div.dir = "auto";
    document.body.appendChild(div);

    setTimeout(() => div.classList.add("closing"), 5000);
    setTimeout(() => div.remove(), 5300);
}

document.body.addEventListener("htmx:responseError", (event: CustomEvent) => {
    const xhr = event.detail.xhr as XMLHttpRequest;
    const responseText = xhr.responseText?.trim();
    showErrorMessage(
        responseText ||
            `Error (${xhr.status}): An unknown server error occurred.`
    );
});

document.body.addEventListener("htmx:sendError", () => {
    showErrorMessage("Network error: Server unreachable or request failed.");
});
