$(document).ready(function() {
    lightbox.option({
        'resizeDuration': 200,
        'wrapAround': true
    });
     
    $(window).scroll(function() {
        let position = $(this).scrollTop();
        
        if (position >= 350) {
            $('.gallery').addClass('change');
        } else {
            $('.gallery').removeClass('change');
        }
    });
});

function searchFunction() {
    var input, filter, div, form, a, i;
    input = document.getElementById('myinput');
    filter = input.value.toUpperCase();
    div = document.getElementById('wrapper');
    form = div.getElementsByTagName('form');

    for(i=0 ; i< form.length; i++){
        a = form[i].getElementsByTagName('a')[0];
        if(a.innerHTML.toUpperCase().indexOf(filter) > -1){
            li[i].style.display = "";
        }

        else{
            li[i].style.display = 'none';
        }
    }
}
