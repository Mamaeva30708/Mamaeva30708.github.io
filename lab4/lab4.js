var Cities = ["Venice", "Rome", "Kailua-Kona"];
for( var i=0; i<Cities.length; i++){
 document.write(Cities[i] + '<br>');
}

function fact(sbj, obj ) {
    return {
        subject: sbj,
        object: obj,
        displayInfo: function() {
            document.write("Random fact " + this.subject + " : " + " I love " + this.object);
        }
    };
};

var funct = fact("about me", "marzipan");
funct.displayInfo();
