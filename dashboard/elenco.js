var apiUrl = 'http://localhost:8080/api/v1';

/*
Chiamata AJAX all'url specificato e relativa funzione da chiamare per gestire il json.
I dati sono ritornati parserizzati tramite $.parseJSON().
*/
$.getJSON(apiUrl + "/dumpsters", function(data) {

    // Creo il bottone uguale per ogni dumpster.
    var btnDetails = '<td><button class="button">Abilita/Disabilita</button></td>';

    $.each(data.dumpsters, function(key, val) {
        var tdId = `<td class="idDumpster"><a href="statisticheBidone.html?id=${val.id}">${val.id}</a></td>`;
        var tdName = '<td>' + val.name + '</td>';
        var tdStatus = '<td class="status">' + val.available + '</td>';
        var tdCountDumps = '<td>' + val.dumps_since_last_emptied + '</td>';
        var tdCurrentWeight = '<td>' + val.current_weight + '</td>';
        var tdLimitWeight = '<td>' + val.weight_limit + '</td>';

        // Riempio la tabella.
        $('#dumpstersTable').append('<tr>' + tdId + tdName + tdCountDumps 
            + tdCurrentWeight + tdLimitWeight + tdStatus + btnDetails + '</tr>');

    });

    document.querySelector('.button').addEventListener("click", function() {
        status = this.parentNode.parentNode.querySelector('.status').innerHTML;
        idDumpster = this.parentNode.parentNode.querySelector('.idDumpster > a').innerHTML;
        status == 'true' ? status = 'false' : status = 'true';
        //this.parentNode.parentNode.querySelector('#status').innerHTML = status;

        // Invio il valore di status
        $.ajax({
            type: "POST",
            url: apiUrl + "/dumpster/" + idDumpster + "/availability",
            contentType: "application/json",
            data: JSON.stringify({available: status})
        });
    });
});

