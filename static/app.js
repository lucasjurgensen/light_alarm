document.addEventListener("DOMContentLoaded", () => {
    const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
    const dayContainer = document.getElementById("day-container");
    let schedules = {};

    // Load all schedules on page load
    const loadAllSchedules = () => {
        fetch(`/api/schedules`)
            .then(response => response.json())
            .then(data => {
                console.log("Fetched the following from /api/schedules:", JSON.stringify(data, null, 2))
                schedules = {};
                data.forEach(schedule => {
                    schedules[schedule.day] = schedule;
                })
                displayAllSchedules(schedules);
            })
            .catch(error => {
                console.error("Error fetching schedules:", error);
                // If fetch fails, display default schedules
                displayAllSchedules(buildDefaultSchedules());
            });
    };

      const buildDefaultSchedules = () => {
          const defaultSchedules = {};
          days.forEach(day => {
              defaultSchedules[day] = { start: 360, end: 540, enabled: true, day: day };
          });
          return defaultSchedules;
      }

    const displayAllSchedules = (loadedSchedules) => {
        schedules = loadedSchedules ? loadedSchedules : buildDefaultSchedules();

        dayContainer.innerHTML = "";
        days.forEach(day => {
            const schedule = schedules[day];
            dayContainer.innerHTML += `
            <div class="day ${schedule.enabled ? '' : 'disabled'}" id="${day}-container">
                <h3>${day}</h3>
                <div class="time-control">
                    <button class="time-button minus-time" data-day="${day}">-</button>
                    <div class="time-box">
                        <span id="${day}-time" contenteditable="true">${minutesToTime(schedule.start)}</span>
                    </div>
                    <button class="time-button plus-time" data-day="${day}">+</button>
                </div>
                <button class="toggle-button ${schedule.enabled ? 'enabled' : ''}" id="${day}-toggle" data-day="${day}"></button>
            </div>
            `;
        });
          // Use event delegation
        dayContainer.addEventListener('click', handleTimeChange);

        days.forEach(day => {
            attachTimeBoxListener(day);
            attachToggleListener(day);
        });
    };

    const handleTimeChange = (event) => {
        if (event.target.classList.contains('time-button')) {
            const day = event.target.dataset.day;
            const isPlus = event.target.classList.contains('plus-time');
            const change = isPlus ? 5 : -5;
            adjustTime(day, change);
        }
    }

    const adjustTime = (day, change) => {
        schedules[day].start += change;
        if (schedules[day].start < 0) {
            schedules[day].start = 0; //Prevent going less than 0
        } else if(schedules[day].start > 1435){
            schedules[day].start = 1435; //Prevent going over 23:59
        }
        const formattedTime = minutesToTime(schedules[day].start);
        document.getElementById(`${day}-time`).textContent = formattedTime;
    };



    const attachTimeBoxListener = (day) => {
        const timeBox = document.getElementById(`${day}-time`);
        timeBox.addEventListener("blur", () => {
            const time = timeBox.textContent;
            const minutes = timeToMinutes(time);
            schedules[day].start = minutes;
            document.getElementById(`${day}-time`).textContent = minutesToTime(minutes);
        });
        timeBox.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault(); // Prevent adding a new line
                timeBox.blur(); // Remove focus from contentEditable
            }
        });
    };


    const attachToggleListener = (day) => {
        const toggleButton = document.getElementById(`${day}-toggle`);
        toggleButton.addEventListener("click", () => {
            schedules[day].enabled = !schedules[day].enabled;
            toggleButton.classList.toggle('enabled', schedules[day].enabled);
            document.getElementById(`${day}-container`).classList.toggle('disabled', !schedules[day].enabled);
        });
    }

    // Save schedule
    document.getElementById("save-button").addEventListener("click", () => {
        const updatedSchedules = [];
        days.forEach(day => {
            updatedSchedules.push({
                day: day,
                start: schedules[day].start,
                end: schedules[day].start + 60,
                enabled: schedules[day].enabled
            });
        });
        console.log("Data immediately before fetch:", JSON.stringify(updatedSchedules));
        fetch("/api/schedules/save", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(updatedSchedules),
        })
            .then(response => {
                if (response.ok) {
                    alert(`Schedules saved successfully!`);
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


    // Load all schedules on page load
    loadAllSchedules();
});

// Helper: Convert minutes from midnight to HH:mm
function minutesToTime(minutes) {
    const hours = Math.floor(minutes / 60).toString().padStart(2, "0");
    const mins = (minutes % 60).toString().padStart(2, "0");
    return `${hours}:${mins}`;
}

function timeToMinutes(time) {
    const [hours, minutes] = time.split(':').map(Number);
    return hours * 60 + minutes;
}