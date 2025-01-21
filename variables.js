
    let socket;
    let keyDownCounter = 0;
    const messageVersion = 1;
    const MessageType = {
        TextMessage: 0,
        GameStateUpdate: 1,
        GameSetup: 2,
        PlayerListUpdate: 3,
        HighscoreUpdate: 4,
    };

    let levelWidth = 50;
    let levelHeight = 50;
    const scale = 8;
    let canvasWidth = levelWidth * scale;
    let canvasHeight = levelHeight * scale;
    let c = document.getElementById('c');

    let input = document.getElementById("input");
    input.value = document.location.href.replace(/https?:\/\//g, "").replace(":8080/", "").replace("/", "");

    let names = ["William Snakespeare",
        "Hissy Elliott",
        "Boa Fett",
        "Noodles",
        "Jeff",
        "Cobra OBrien",
        "Snake Gyllenhaal",
        "Conan The Boabarian",
        "Reese Slitherspoon",
        "Sneaky Snake"];

    document.getElementById("snakeName").value = names[Math.floor(Math.random() * names.length)];

    c.width = canvasWidth;
    c.height = canvasHeight;
    let ctx = c.getContext('2d');
    ctx.fillStyle = "#000";

    let keysDown = {};
    let keysHit = {};
