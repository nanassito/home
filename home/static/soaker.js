document.querySelectorAll("#burst button").forEach(function(btn) {
    btn.addEventListener("click", async function() {
        btn.disabled = true
        try {
            await fetch("/api/valve/activate", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    area: btn.dataset.target,
                    duration_sec: 10,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        btn.disabled = false
    })
})