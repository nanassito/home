document.querySelectorAll(".settings button.btn-activate[data-control='auto']").forEach(function(btn) {
    var target = document.querySelector("#" + btn.dataset.context + " #control-auto")
    btn.addEventListener("click", async function() {
        try {
            await fetch("/api/room", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    "room": btn.dataset.room,
                    "min_temp": target.querySelector("#in-min-temp").value * 1,
                    "max_temp": target.querySelector("#in-max-temp").value * 1,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        $("#" + btn.dataset.context).modal("hide");
        window.location.reload();
    })
});

document.querySelectorAll(".settings button.btn-activate[data-control='app']").forEach(function(btn) {
    var target = document.querySelector("#" + btn.dataset.context + " #control-app")
    btn.addEventListener("click", async function() {
        try {
            await fetch("/api/hvac/app", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    "hvac": btn.dataset.hvac,
                    "mode": target.querySelector("#ddbtn-mode").value,
                    "fan": target.querySelector("#ddbtn-fan").value,
                    "target_temp": target.querySelector("#in-temp").value * 1,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        $("#" + btn.dataset.context).modal("hide");
        window.location.reload();
    })
});

document.querySelectorAll(".settings button.btn-activate[data-control='remote']").forEach(function(btn) {
    var target = document.querySelector("#" + btn.dataset.context + " #control-remote")
    btn.addEventListener("click", async function() {
        try {
            await fetch("/api/hvac/remote", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    "hvac": btn.dataset.hvac,
                })
            });
        } catch (err) {
            window.location.reload();
        }
        $("#" + btn.dataset.context).modal("hide");
        window.location.reload();
    })
});