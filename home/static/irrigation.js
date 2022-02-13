document.querySelectorAll("#start").forEach(function(btn) {
    btn.addEventListener("click", async function() {
        btn.disabled = true
        try {
            await fetch("/api/valve/activate", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    area: document.getElementById("zone").value,
                    duration_sec: document.getElementById("duration").value * 60,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        btn.disabled = false
    })
})