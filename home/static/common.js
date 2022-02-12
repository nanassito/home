document.querySelectorAll("input[role=switch]").forEach(function(checkbox) {
    checkbox.addEventListener("click", async function() {
        checkbox.disabled = true
        try {
            await fetch("/api/feature_flag", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    target: checkbox.dataset.target,
                    enabled: checkbox.checked,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        checkbox.disabled = false
    })
})