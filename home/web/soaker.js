document.querySelectorAll("#weapons .save").forEach(function (button) {
    button.addEventListener("click", async function () {
        try {
            await fetch("/api/weapons/soaker", {
                method: "POST",
                redirect: "follow",  // So that we don't need to enforce consistency in the UI.
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ enabled: document.querySelector("#weapons input").checked })
            });
        } catch (err) {
            alert("failed to save changes: " + err)
        }
    })
})

document.addEventListener("DOMContentLoaded", async () => {
    var soaker = await (await fetch("/api/weapons/soaker")).json()
    document.querySelector("#soaker_enabled").checked = soaker.enabled
})