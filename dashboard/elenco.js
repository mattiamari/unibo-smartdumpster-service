const apiUrl = 'http://smartdumpster.mattiamari.me/api/v1';

/*
Chiamata AJAX all'url specificato e relativa funzione da chiamare per gestire il json.
I dati sono ritornati parserizzati tramite $.parseJSON().
*/
fetch(apiUrl + "/dumpsters").then(response => response.json()).then(data => {
    // Creo il bottone uguale per ogni dumpster.
    const table = document.querySelector('#dumpstersTable');

    for (const dumpster of data.dumpsters) {
        const row = document.createElement('tr');
        row.innerHTML =
            `<td class="idDumpster"><a href="stats.html?id=${dumpster.id}">${dumpster.id}</a></td>
            <td>${dumpster.name}</td>
            <td>${dumpster.dumps_since_last_emptied}</td>
            <td>${dumpster.current_weight}</td>
            <td>${dumpster.weight_limit}</td>
            <td class="status">${dumpster.available}</td>
            <td><button class="btnAvailability">Abilita/Disabilita</button></td>`;

        table.append(row);
    }

    document.querySelectorAll('.btnAvailability').forEach(e => {
        e.addEventListener("click", function() {
            const idDumpster = this.parentNode.parentNode.querySelector('.idDumpster > a').innerHTML;
            let status = this.parentNode.parentNode.querySelector('.status').innerHTML;
            status = status == 'true' ? false : true; // faccio lo stato opposto
            //this.parentNode.parentNode.querySelector('#status').innerHTML = status;

            // Invio il valore di status
            fetch(apiUrl + "/dumpster/" + idDumpster + "/availability", {
                method: "POST",
                mode: "cors",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({available: status})
            });
        });
    });
});
