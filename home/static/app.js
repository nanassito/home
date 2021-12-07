document.querySelectorAll(".feature_flag_switch").forEach(function (checkbox) {
    checkbox.addEventListener("click", async function () {
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
            alert("failed to save changes: " + err)
        }
        document.location.pathname = "/"  // Force reload everything
    })
})