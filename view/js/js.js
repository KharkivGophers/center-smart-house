function parseURLParams(url) {
    var queryStart = url.indexOf("?") + 1,
        queryEnd = url.indexOf("#") + 1 || url.length + 1,
        query = url.slice(queryStart, queryEnd - 1),
        pairs = query.replace(/\+/g, " ").split("&"),
        parms = {}, i, n, v, nv;

    if (query === url || query === "") return;

    for (i = 0; i < pairs.length; i++) {
        nv = pairs[i].split("=", 2);
        n = decodeURIComponent(nv[0]);
        v = decodeURIComponent(nv[1]);

        if (!parms.hasOwnProperty(n)) parms[n] = [];
        parms[n].push(nv.length === 2 ? v : null);
    }
    return parms;
}

function sendDevConfigFreq(id, collectFreq, sendFreq) {
    var xhr = new XMLHttpRequest();
    var url = "/devices/" + id + "/config";
    xhr.open("PATCH", url, true);
    xhr.setRequestHeader("Content-type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            alert("Data have been delivered!");
        }
    };

    var config = JSON.stringify(
        {
            "collectFreq": collectFreq,
            "sendFreq": sendFreq
        });

    xhr.send(config);
}

function sendDevConfigTurnedOn(id, turnedOn) {
    var xhr = new XMLHttpRequest();
    var url = "/devices/" + id + "/config";
    xhr.open("PATCH", url, true);
    xhr.setRequestHeader("Content-type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            alert("Data have been delivered!");
        }
    };

    var config = JSON.stringify(
        {
            "turnedOn": turnedOn
        });

    xhr.send(config);
}

function sendDevConfigStreamOn(id, streamOn) {
    var xhr = new XMLHttpRequest();
    var url = "/devices/" + id + "/config";
    xhr.open("PATCH", url, true);
    xhr.setRequestHeader("Content-type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            alert("Data have been delivered!");
        }
    };

    var config = JSON.stringify(
        {
            "streamOn": streamOn
        });

    xhr.send(config);
}

function setDevDataFields(obj) {
    document.getElementById('devType').value = obj["meta"]["type"];
    document.getElementById('devName').value = obj["meta"]["name"];
}

function setDevConfigFields(obj) {
    if (obj["turnedOn"]) {
        document.getElementById('turnedOnBtn').innerHTML = "On";
        document.getElementById('turnedOnBtn').className = "btn btn-success";
    } else {
        document.getElementById('turnedOnBtn').innerHTML = "Off";
        document.getElementById('turnedOnBtn').className = "btn btn-danger";
    }

    document.getElementById('collectFreq').value = obj["collectFreq"];
    document.getElementById('sendFreq').value = obj["sendFreq"];

    if (obj["streamOn"]) {
        document.getElementById('streamOnBtn').innerHTML = "On";
        document.getElementById('turnedOnBtn').className = "btn btn-success";
    } else {
        document.getElementById('streamOnBtn').innerHTML = "Off";
        document.getElementById('turnedOnBtn').className = "btn btn-danger";
    }
}

function printFridgeChart(obj) {
    //chart
    Highcharts.setOptions({
        global: {
            useUTC: false
        }
    });

    // Create the chart
    Highcharts.stockChart('container', {
        chart: {
            events: {
                load: function () {
                    // set up the updating of the chart each second

                    /*var series = this.series[0];
                     setInterval(function () {
                     var x = (new Date()).getTime(), // current time
                     y = Math.round(Math.random() * 100);
                     series.addPoint([x, y], true, true);
                     }, 1000);
                     */
                }
            }
        },

        rangeSelector: {
            buttons: [{
                count: 1,
                type: 'minute',
                text: '1M'
            }, {
                count: 5,
                type: 'minute',
                text: '5M'
            }, {
                type: 'all',
                text: 'All'
            }],
            inputEnabled: false,
            selected: 0
        },

        exporting: {
            enabled: false
        },

        series: [{
            name: 'TempCam1',
            data: (function () {
                var data = [];
                for (var i = 0; i < obj["data"]["TempCam1"].length; ++i) {
                    data.push({
                        x: parseInt(obj["data"]["TempCam1"][i].split(':')[0]),
                        y: parseFloat(obj["data"]["TempCam1"][i].split(':')[1])
                    });
                }
                return data;
            }())
        }, {
            name: 'TempCam2',
            data: (function () {
                var data = [];
                for (var i = 0; i < obj["data"]["TempCam2"].length; ++i) {
                    data.push({
                        x: parseInt(obj["data"]["TempCam2"][i].split(':')[0]),
                        y: parseFloat(obj["data"]["TempCam2"][i].split(':')[1])
                    });
                }
                return data;
            }())
        }]
    })
}

