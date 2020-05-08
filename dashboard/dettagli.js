const urlParams = new URLSearchParams(window.location.search);
const id = urlParams.get('id');

var apiUrl = "http://localhost:8080/api/v1/dumpster/" + id;


var dataNumDumps = [];
var dataWeight = [];
var dataDumpsTime = [];
var dataPoints = [];
var arrayWeightLimit = [];
var dumpsType = {};

$.getJSON(apiUrl, function(data) {
    weight_limit = data.dumpster["weight_limit"];
    dump_history = data.dumpster["dump_history"];

    weight_history = data.dumpster["weight_history"];


    $.each(dump_history, function(key, val) {
        dumpsType[val.dump_type] = (dumpsType[val.dump_type] || 0) + 1;
    });
    

    $.each(weight_history, function(key, val) {
        
        dataNumDumps.push(val.dumps_since_last_emptied);
        dataWeight.push(val.weight);

        arrayWeightLimit.push(weight_limit);

        var pos = val.created_at.lastIndexOf(".");

        dataDumpsTime.push(val.created_at.slice(0, pos));

        dataPoints.push({x: val.created_at.slice(0, pos), y: val.weight });
    });



    var ctx1 = document.getElementById('myChart1');

    var myChart = new Chart(ctx1, {
        type: 'line',
        data: {
            labels: dataNumDumps,
            datasets: [{ 
                data: dataWeight,
                label: "Peso",
                borderColor: "#3e95cd",
                fill: false,
            }, {
                label: "Capienza massima",
                borderColor: "rgb(0, 0, 0)",
                data: arrayWeightLimit,
                //borderDash: [0,50],
                fill: false,  
            }]
        },
        options: {
            title: {
                display: true,
                text: 'Grafico dei depositi'
            }
        }
    });

    var ctx2 = document.getElementById('myChart2');

    var my = new Chart(ctx2, {
        type: 'line',
        data: {
            labels: dataDumpsTime,

            datasets: [{
                label: 'Peso',
                yAxisID: 'A',
                data: dataWeight,
                borderColor: "#ff6384",
                fill: false
            }, {
                label: 'Numero depositi',
                yAxisID: 'B',
                borderColor: 'rgb(254, 217, 118)',
                data: dataNumDumps,
                fill: false
            }]   
        },
        options: {
            title: {
                display: true,
                text: 'Grafico temporale'
            },
            scales: {
                yAxes: [{
                    id: 'A',
                    type: 'linear',
                    position:'left' 
                }, {
                    id: 'B',
                    type: 'linear',
                    position: 'right'
                }],
                xAxes: [{
                    type: 'time',
                    time: {
                        displayFormats: {
                            millisecond: 'h:mm:ss.SSS'
                        }
                    }
                }]
            }
        }
    });

    var dumpsTypeValue = []
    var dumpsTypeOccurences = [];

    for(var index in dumpsType) {
        dumpsTypeOccurences.push(dumpsType[index]);
        dumpsTypeValue.push(index)
    }
    
    var ctx3 = document.getElementById('myChart3');

    var a = new Chart(ctx3, {
        type: 'pie',
        data: {
            datasets: [{
                data: dumpsTypeOccurences,
                backgroundColor: ["rgb(254, 217, 118)", "rgb(8, 88, 158)","rgb(150, 150, 150)"],
            }],      
            labels: dumpsTypeValue, 
        },
        options: {
            title: {
                display: true,
                text: 'Distribuzione tipologia di rifiuti'
            }
        },
    });
});