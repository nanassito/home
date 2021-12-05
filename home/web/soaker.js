document.querySelectorAll("#weapons input").forEach(function (checkbox) {
    checkbox.addEventListener("click", async function () {
        try {
            await fetch("/api/weapons/soaker", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ enabled: checkbox.checked })
            });
        } catch (err) {
            alert("failed to save changes: " + err)
        }
        document.location.pathname = "/"  // Force reload everything
    })
})