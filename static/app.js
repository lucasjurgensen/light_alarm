document.addEventListener("DOMContentLoaded", () => {
    const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];

    // Determine current day and tomorrow
    const todayIndex = new Date().getDay() - 1; // JS starts Sunday as 0
    const today = days[(todayIndex + 7) % 7]; // Handle edge cases like Sunday
    const tomorrow = days[(todayIndex + 1) % 7];

    // Populate dropdown with tomorrow as the default selected day
    const daySelector = document.getElementById("day-selector");
    daySelector.innerHTML = `
        <option value="${tomorrow}" selected>Tomorrow (${tomorrow})</option>
        <option value="${today}">Today (${today})</option>
        ${days.filter(day => day !== today && day !== tomorrow)
            .map(day => `<option value="${day}">${day}</option>`)
            .join("")}
    `;

    // Load and display tomorrow's schedule as the default
    const dayContainer = document.getElementById("day-container");
    let schedules = {}; // Cache schedules to avoid repeated API calls

    const loadDaySchedule = (day) => {
        if (schedules[day]) {
            displayDaySchedule(day, schedules[day]);
        } else {
            fetch(`/api/schedules`)
                .then(response => response.json())
                .then(data => {
                    schedules = data; // Cache all schedules
                    displayDaySchedule(day, schedules[day]);
                });
        }
    };

    const displayDaySchedule = (day, schedule) => {
        dayContainer.innerHTML = `
            <div class="day">
                <h3>${day}</h3>
                <div class="slider-container">
                    <label class="slider-label" for="${day}-start">Start Time:</label>
                    <input type="range" id="${day}-start" min="0" max="1439" value="${schedule?.start || 420}">
                    <div class="slider-value" id="${day}-start-value">${minutesToTime(schedule?.start || 420)}</div>
                </div>
                <div class="slider-container">
                    <label class="slider-label" for="${day}-end">End Time:</label>
                    <input type="range" id="${day}-end" min="0" max="1439" value="${schedule?.end || 480}">
                    <div class="slider-value" id="${day}-end-value">${minutesToTime(schedule?.end || 480)}</div>
                </div>
            </div>
        `;
        attachSliderListeners(day);
    };

    const attachSliderListeners = (day) => {
        document.querySelectorAll(`#${day}-start, #${day}-end`).forEach(slider => {
            slider.addEventListener("input", (e) => {
                const value = e.target.value;
                const formattedTime = minutesToTime(value);
                document.getElementById(e.target.id + "-value").textContent = formattedTime;
            });
        });
    };

    // Save schedule
    document.getElementById("save-button").addEventListener("click", () => {
        const selectedDay = daySelector.value;
        const updatedSchedule = {
            [selectedDay]: {
                start: parseInt(document.getElementById(`${selectedDay}-start`).value, 10),
                end: parseInt(document.getElementById(`${selectedDay}-end`).value, 10),
            },
        };

        // Send to server
        fetch("/api/schedules/save", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(updatedSchedule),
        })
            .then(response => {
                if (response.ok) {
                    alert(`${selectedDay}'s schedule saved successfully!`);
                    schedules[selectedDay] = updatedSchedule[selectedDay]; // Update local cache
                } else {
                    alert("Failed to save schedule.");
                }
            })
            .catch(error => console.error("Error saving schedule:", error));
    });


    // Save schedule
    document.getElementById("trigger-test").addEventListener("click", () => {
        // Send to server
        fetch("/api/test-lights", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
        })
            .then(response => {
                if (response.ok) {
                    alert(`Triggered light test!`);
                } else {
                    alert("Failed to trigger light test.");
                }
            })
            .catch(error => console.error("Error triggering light test:", error));
    });

    // Load tomorrow's schedule on page load
    daySelector.addEventListener("change", () => loadDaySchedule(daySelector.value));
    loadDaySchedule(tomorrow);
});

// Helper: Convert minutes from midnight to HH:mm
function minutesToTime(minutes) {
    const hours = Math.floor(minutes / 60).toString().padStart(2, "0");
    const mins = (minutes % 60).toString().padStart(2, "0");
    return `${hours}:${mins}`;
}