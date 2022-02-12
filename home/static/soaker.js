document.querySelectorAll("#burst button").forEach(function(btn) {
    btn.addEventListener("click", async function() {
        btn.disabled = true
        try {
            await fetch("/api/valve/burst", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    area: btn.dataset.target,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        btn.disabled = false
    })
})