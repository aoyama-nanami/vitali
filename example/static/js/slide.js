function recenter(){
    $("#slide_wrapper").css("padding-left", (($(window).width()-1024)/2)+"px")
    $("#slide_wrapper").css("padding-right", (($(window).width()-1024)/2)+"px")
    $("#slide_wrapper").css("margin-top", ($(window).height()-$("#slide_wrapper").height())/2+"px")
    $("#slide_wrapper").css("visibility", "visible")
}

$(document).ready(recenter)
$("img").load(recenter)

function show_css(){
    $("#raw").hide()
    $("#css").show()
    $("#raw_tab").removeClass("active")
    $("#css_tab").addClass("active")
}

function show_raw(){
    $("#css").hide()
    $("#raw").show()
    $("#css_tab").removeClass("active")
    $("#raw_tab").addClass("active")
}

$("#edit_modal").on("shown.bs.modal", function(){
    var textarea = $(this).find("textarea")[0]
    textarea.selectionStart = textarea.value.length
    textarea.selectionEnd = textarea.value.length
})

$("body").bind("touchmove", function(e){
    e.preventDefault()
})

function next_page() {
    if(parseInt($.cookie("page"))+1 > total_pages) {
        return
    }
    location.href = parseInt($.cookie("page"))+1
}

function prev_page() {
    if($.cookie("page") == "1") {
        return
    }
    location.href = parseInt($.cookie("page"))-1
}

$(window).on("swipeleft", next_page)
$(window).on("swiperight", prev_page)

function delete_page() {
    $.ajax({
        type: "DELETE",
    }).done(function(){
        if(parseInt($.cookie("page")) == total_pages) {
            prev_page()
        }
        location.href = ""
    })
}

function toggle_remote() {
    if($.cookie("remote") == "true") {
        $.removeCookie("remote")
    }
    else {
        $.cookie("remote", "true")
    }
    location.href = ""
}

$("#left_half").click(prev_page)
$("#right_half").click(next_page)

$(document).keydown(function(e){
    if($("#edit_modal").css("display") != "none") {
        return
    }
    if(e.which == 39) {
        next_page()
    }
    else if(e.which == 37) {
        prev_page()
    }
    else if(e.which == 36) {
        location.href = '1'
    }
    else if(e.which == 35) {
        location.href = total_pages
    }
})

if($.cookie("remote") == "true") {
    $(".glyphicon-phone").css("color", "red")
    $.post("/progress/"+slide_name, {page: $.cookie("page")})
}
else {
    $.get("/progress/"+slide_name).done(function(res){
        location.href = res
    })
}

if($.cookie("create") == "create") {
    $("#edit_modal").modal({})
    $.removeCookie("create", {path: location.pathname})
}
