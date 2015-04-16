importPackage(Packages.java.util);
importPackage(Packages.com.dv.gtusach.server.common);
importPackage(Packages.org.jsoup);
importPackage(Packages.org.jsoup.nodes);
importPackage(Packages.org.jsoup.select);

function getBatchSize() {
	return 50;
}

function getDelayTimeSec() {
	return 10;
}

function getChapterTitle(html, formatHtml) {
  var chapterTitle = "";
  var title1 = "";
  var title2 = "";
  
  var index = 0;
  var i, j;
  while ((index = html.indexOf("post_message_", index)) != -1) {    
    i = html.indexOf(" onClick=\"", index);
    index++;
    var index2 = html.indexOf("a href=", index);    
    if (index2 == -1) {
      index2 = index + 5000;
    }
    
    if ((i != -1) && (i < index2)) {
      i = html.indexOf("<br", i);
      j = html.indexOf("</div>", i);
      if (j != -1) {
        if (j > index2) {
          j = index2
        }
        // get the title
        var str = getDefaultChapterTitle(html.substring(index, i+500), html.substring(index, i+500));
        if (str != "") {
          if (title1 == "") {
            title1 = str;
          } else {
            title2 = str;
          }  				  				      
        }                
      }       
    } else {
      continue;
    }
  }
  
  chapterTitle = title1 + "/" + title2;
  if (chapterTitle.length > 64) {
    chapterTitle = chapterTitle.substring(0, 64);
  }
  return chapterTitle;
}

function getNextPageUrl(targetURL, currentPageURL, html) { 
    var result = null;
    var doc = Jsoup.parse(html);
    var body = doc.getElementsByTag("body").first();
    var list = body.getElementsByTag("a");
    for (var i=0; i<list.size(); i++) {
      var elm = list.get(i);
      if (elm.attr("rel").equals("next")) {
      	result = elm.attr("href");
      	break;
      }
    }
    if (result != null) {
    	if (result.startsWith("showthread")) {
    		result = "forum/" + result;
    	} else if (result.startsWith("/showthread")) {
    		result = "/forum" + result;    		
    	}
    }
    return result;
}


function extractChapterHtml(targetStr, requestStr, html) {
  var textStr = "";  	  
  var index = 0;
  var i, j;
  while ((index = html.indexOf("post_message_", index)) != -1) {    
    i = html.indexOf(" onClick=\"", index);
    index++;
    var index2 = html.indexOf("a href=", index);    
    if (index2 == -1) {
      index2 = index + 5000;
    }
    
    if ((i != -1) && (i < index2)) {
      i = html.indexOf("<br", i);
      j = html.indexOf("</div>", i);
      if (j != -1) {
        if (j > index2) {
          j = index2
        }
        // get the title
        var str = getDefaultChapterTitle(html.substring(index, i+500), html.substring(index, i+500));
        if (str != "") {
          textStr += "<p style=\"page-break-before: always\">" + str + "<br/><br/>" + html.substring(i, j);  						
        } else {
          textStr += "<p style=\"page-break-before: always\">" + html.substring(i, j);  						
        }               
      }       
    } else {
      continue;
    }
  }
  
  var chapterHtml = null;
  if (textStr.length > 0) {
    chapterHtml = new ChapterHtml();
    index = context.getBookTemplate().indexOf("</body>");
    chapterHtml.setHtml(context.getBookTemplate().substring(0, index-1) + textStr + "</body></html>");           
  }
  
  return chapterHtml;
}
