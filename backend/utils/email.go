package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

const EMAIL_TEMPLATE_INVITATION = `<!doctype html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">

<head>
<meta charset="utf-8" />
<meta content="width=device-width" name="viewport" />
<meta content="IE=edge" http-equiv="X-UA-Compatible" />
<meta name="x-apple-disable-message-reformatting" />
<meta content="telephone=no,address=no,email=no,date=no,url=no" name="format-detection" />
<title>Invite User to Project</title>
<!--[if mso]>
            <style>
                * {
                    font-family: sans-serif !important;
                }
            </style>
        <![endif]-->
<!--[if !mso]><!-->
<!-- <![endif]-->
<link href="https://fonts.googleapis.com/css?family=Inter:600" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:400" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:500" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Poppins:700" rel="stylesheet" type="text/css">
<style>
html {
    margin: 0 !important;
    padding: 0 !important;
}

* {
    -ms-text-size-adjust: 100%;
    -webkit-text-size-adjust: 100%;
}
td {
    vertical-align: top;
    mso-table-lspace: 0pt !important;
    mso-table-rspace: 0pt !important;
}
a {
    text-decoration: none;
}
img {
    -ms-interpolation-mode:bicubic;
}
@media only screen and (min-device-width: 320px) and (max-device-width: 374px) {
    u ~ div .email-container {
        min-width: 320px !important;
    }
}
@media only screen and (min-device-width: 375px) and (max-device-width: 413px) {
    u ~ div .email-container {
        min-width: 375px !important;
    }
}
@media only screen and (min-device-width: 414px) {
    u ~ div .email-container {
        min-width: 414px !important;
    }
}

</style>
<!--[if gte mso 9]>
        <xml>
            <o:OfficeDocumentSettings>
                <o:AllowPNG/>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
        <![endif]-->
<style>
@media only screen and (max-device-width: 898px), only screen and (max-width: 898px) {

    .eh {
        height:auto !important;
    }
    .desktop {
        display: none !important;
        height: 0 !important;
        margin: 0 !important;
        max-height: 0 !important;
        overflow: hidden !important;
        padding: 0 !important;
        visibility: hidden !important;
        width: 0 !important;
        min-height: 0 !important;
    }
    .mobile {
        display: block !important;
        width: auto !important;
        height: auto !important;
        float: none !important;
    }
    .email-container {
        width: 100% !important;
        margin: auto !important;
    }
    .stack-column,
    .stack-column-center {
        display: block !important;
        width: 100% !important;
        max-width: 100% !important;
        direction: ltr !important;
    }
    .wid-auto {
        width:auto !important;
    }

    .table-w-full-mobile {
        width: 100%;
    }
    .full-button {
        width: 100% !important;
    }

    .text-35336193 {font-size:20px !important;}.text-42988535 {font-size:16px !important;}.text-46290937 {font-size:16px !important;}.text-71124296 {font-size:12px !important;}.text-05283882 {font-size:12px !important;}.text-73760468 {font-size:12px !important;}.text-60840158 {font-size:14px !important;}
    .pt-33891614 {padding-top:20px !important;}.pb-28658058 {padding-bottom:20px !important;}.pl-12855736 {padding-left:20px !important;}.pr-06317584 {padding-right:20px !important;}.pt-56942309 {padding-top:24px !important;}.pb-55295092 {padding-bottom:24px !important;}.pl-59418156 {padding-left:14px !important;}.pr-45511306 {padding-right:14px !important;}.pt-81133467 {padding-top:16px !important;}.pb-39479741 {padding-bottom:16px !important;}.pl-39842558 {padding-left:16px !important;}.pr-55630940 {padding-right:16px !important;}.pt-48740625 {padding-top:16px !important;}.pb-75138081 {padding-bottom:16px !important;}.pl-76362343 {padding-left:16px !important;}.pr-49067704 {padding-right:16px !important;}

    .mobile-center {
        text-align: center;
    }

    .mobile-center > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-left {
        text-align: left;
    }

    .mobile-left > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-right {
        text-align: right;
    }

    .mobile-right > table {
        display: inline-block;
        vertical-align: inherit;
    }

}
@media (prefers-color-scheme:dark){ .fniqkbzkco {background-color:#161616 !important} .vuzgfmogon {background-color:#333232 !important} .gnxtinfpfx {color: #ffffff !important} .kxlbjxevht {color: #BFBFBF !important} .wvujawwxnf {color: #BFBFBF !important} .wqdggqjquw {color: #ffffff !important} } 
</style>
</head>

<body width="100%" style="background-color:#dedede;margin:0;padding:0!important;mso-line-height-rule:exactly;">
<div style="background-color:#dedede">
<!--[if gte mso 9]>
                                            <v:background xmlns:v="urn:schemas-microsoft-com:vml" fill="t">
                                            <v:fill type="tile" color="#dedede"/>
                                            </v:background>
                                            <![endif]-->
<table width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td valign="top" align="center">
<table bgcolor="#000000" style="margin:0 auto;" align="center" id="brick_container" cellspacing="0" cellpadding="0" border="0" width="899" class="email-container">
<tr>
<td width="899">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="899" align="center" class="fniqkbzkco pt-33891614 pb-28658058 pl-12855736 pr-06317584" style="background-color:#000000;   padding-left:120px; padding-right:120px;" bgcolor="#000000">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="414" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="218" align="center"><img src="https://media.marka-img.com/983d3508/1jSQ0zaOTtFzujJaHOCxgTe3pDPAK2.png" width="218" border="0" style="max-width:218px; width: 100%;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="715">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="715" align="center" class="vuzgfmogon pt-56942309 pb-55295092 pl-59418156 pr-45511306" style="vertical-align: middle; background-color:#101014; border-radius:12px;  box-shadow: 0px 0px 8px 0px rgba(0, 0, 0, 0.10000000149011612);" bgcolor="#101014">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-81133467 pb-39479741 pl-39842558 pr-55630940" style="vertical-align: middle;   padding-left:64px; padding-right:64px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="64" align="center"><img src="https://media.marka-img.com/983d3508/aNOnLC4WFXoLh8MYHOTQsRctpuMPty.png" width="64" border="0" style="min-width:64px; width:64px;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:24px; height:24px; font-size:24px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<div style="line-height:normal;text-align:center;"><span class="text-35336193 gnxtinfpfx" style="color:#ffffff;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:28px;line-height:normal;text-align:center;">John Doe Invited you to Join Project</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-42988535 kxlbjxevht" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Hi Krishna,<br><br>This is an invitation from John Doe to join the project “TRUVIZ INC” to collaborate. Click the button below to set up your account and get started. </span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="right" style="vertical-align: middle; height:45px; background-color:#11e59e; border-radius:8px; border:1.340000033378601px solid #11e59e; box-shadow: 0px 1.340000033378601px 2.680000066757202px 0px rgba(16, 24, 40, 0.05000000074505806);" bgcolor="#11e59e">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle; height:45px;   padding-left:24px; padding-right:24px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;">
<div style="line-height:normal;text-align:center;"><span style="color:#000000;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:18px;line-height:normal;text-align:center;"><a style="color:#000000;text-decoration:none;" href="InvitationURL" target="_blank">Accept Invitation</a></span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-46290937 wvujawwxnf" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">If you have any questions, Please reach out to the Invitee. </span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-48740625 pb-75138081 pl-76362343 pr-49067704 mobile-center" style="vertical-align: middle;  ">
<table border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/privacy-policy">
<span class="text-71124296">Privacy Policy</span></a></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span style="color:#c0c5cd;font-weight:700;font-family:Poppins,Arial,sans-serif;font-size:16px;line-height:normal;text-align:center;">•</span></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/">
<span class="text-73760468">Contact Us</span></a></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.linkedin.com/company/trugen-ai/"><img src="https://media.marka-img.com/983d3508/xHxxQtmINpVmiAvX3u5wte6R5TRyJk.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
<td style="width:16px; min-width:16px;" width="16">&#8202;</td>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.youtube.com/@trugen_ai"><img src="https://media.marka-img.com/983d3508/0mOGhLi7ep0WZvo8obpBcx1k9sMKz3.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span class="text-60840158 wqdggqjquw" style="color:#ffffff;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;">Copyright © Trugen.ai</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</div>
</body>

</html>`
const EMAIL_TEMPLATE_SUBSCRIPTION = `<!doctype html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">

<head>
<meta charset="utf-8" />
<meta content="width=device-width" name="viewport" />
<meta content="IE=edge" http-equiv="X-UA-Compatible" />
<meta name="x-apple-disable-message-reformatting" />
<meta content="telephone=no,address=no,email=no,date=no,url=no" name="format-detection" />
<title>Your Growth Plan activated</title>
<!--[if mso]>
            <style>
                * {
                    font-family: sans-serif !important;
                }
            </style>
        <![endif]-->
<!--[if !mso]><!-->
<!-- <![endif]-->
<link href="https://fonts.googleapis.com/css?family=Inter:600" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:400" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:700" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:500" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Poppins:700" rel="stylesheet" type="text/css">
<style>
html {
    margin: 0 !important;
    padding: 0 !important;
}

* {
    -ms-text-size-adjust: 100%;
    -webkit-text-size-adjust: 100%;
}
td {
    vertical-align: top;
    mso-table-lspace: 0pt !important;
    mso-table-rspace: 0pt !important;
}
a {
    text-decoration: none;
}
img {
    -ms-interpolation-mode:bicubic;
}
@media only screen and (min-device-width: 320px) and (max-device-width: 374px) {
    u ~ div .email-container {
        min-width: 320px !important;
    }
}
@media only screen and (min-device-width: 375px) and (max-device-width: 413px) {
    u ~ div .email-container {
        min-width: 375px !important;
    }
}
@media only screen and (min-device-width: 414px) {
    u ~ div .email-container {
        min-width: 414px !important;
    }
}

</style>
<!--[if gte mso 9]>
        <xml>
            <o:OfficeDocumentSettings>
                <o:AllowPNG/>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
        <![endif]-->
<style>
@media only screen and (max-device-width: 898px), only screen and (max-width: 898px) {

    .eh {
        height:auto !important;
    }
    .desktop {
        display: none !important;
        height: 0 !important;
        margin: 0 !important;
        max-height: 0 !important;
        overflow: hidden !important;
        padding: 0 !important;
        visibility: hidden !important;
        width: 0 !important;
        min-height: 0 !important;
    }
    .mobile {
        display: block !important;
        width: auto !important;
        height: auto !important;
        float: none !important;
    }
    .email-container {
        width: 100% !important;
        margin: auto !important;
    }
    .stack-column,
    .stack-column-center {
        display: block !important;
        width: 100% !important;
        max-width: 100% !important;
        direction: ltr !important;
    }
    .wid-auto {
        width:auto !important;
    }

    .table-w-full-mobile {
        width: 100%;
    }
    .full-button {
        width: 100% !important;
    }

    .text-34626036 {font-size:20px !important;}.text-20598375 {font-size:16px !important;}.text-28942147 {font-size:16px !important;}.text-58075878 {font-size:16px !important;}.text-30287625 {font-size:16px !important;}.text-81430652 {font-size:12px !important;}.text-67601188 {font-size:12px !important;}.text-93263918 {font-size:12px !important;}.text-80601861 {font-size:14px !important;}
    .pt-99147577 {padding-top:20px !important;}.pb-48112541 {padding-bottom:20px !important;}.pl-17559610 {padding-left:20px !important;}.pr-11293172 {padding-right:20px !important;}.pt-92921938 {padding-top:24px !important;}.pb-93126178 {padding-bottom:24px !important;}.pl-80960861 {padding-left:14px !important;}.pr-42946016 {padding-right:14px !important;}.pt-90589984 {padding-top:16px !important;}.pb-01097341 {padding-bottom:16px !important;}.pl-59594197 {padding-left:16px !important;}.pr-99538240 {padding-right:16px !important;}.pt-02060215 {padding-top:16px !important;}.pb-55703351 {padding-bottom:16px !important;}.pl-20314444 {padding-left:16px !important;}.pr-93347059 {padding-right:16px !important;}

    .mobile-center {
        text-align: center;
    }

    .mobile-center > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-left {
        text-align: left;
    }

    .mobile-left > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-right {
        text-align: right;
    }

    .mobile-right > table {
        display: inline-block;
        vertical-align: inherit;
    }

}
@media (prefers-color-scheme:dark){ .gcwtnqolwl {background-color:#161616 !important} .lmiczbiodn {background-color:#333232 !important} .qvhllyhyni {color: #ffffff !important} .ykmswwraim {color: #BFBFBF !important} .mltdnzchhf {color: #BFBFBF !important} .cyjptgemve {color: #BFBFBF !important} .wcokmlniwt {color: #BFBFBF !important} .iaeczfjgoe {color: #BFBFBF !important} .nlktnbavtq {color: #BFBFBF !important} .gtblrunjln {color: #BFBFBF !important} .grxamfkixj {color: #BFBFBF !important} .ipgtmngmac {color: #BFBFBF !important} .rnsxjemugy {color: #BFBFBF !important} .xrjifbgfyh {color: #BFBFBF !important} .bdrheuatzh {color: #ffffff !important} } 
</style>
</head>

<body width="100%" style="background-color:#dedede;margin:0;padding:0!important;mso-line-height-rule:exactly;">
<div style="background-color:#dedede">
<!--[if gte mso 9]>
                                            <v:background xmlns:v="urn:schemas-microsoft-com:vml" fill="t">
                                            <v:fill type="tile" color="#dedede"/>
                                            </v:background>
                                            <![endif]-->
<table width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td valign="top" align="center">
<table bgcolor="#000000" style="margin:0 auto;" align="center" id="brick_container" cellspacing="0" cellpadding="0" border="0" width="899" class="email-container">
<tr>
<td width="899">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="899" align="center" class="gcwtnqolwl pt-99147577 pb-48112541 pl-17559610 pr-11293172" style="background-color:#000000;   padding-left:120px; padding-right:120px;" bgcolor="#000000">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="414" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="218" align="center"><img src="https://media.marka-img.com/983d3508/O1dYTAD0zhAJfbbM2WJreEOo9owvIK.png" width="218" border="0" style="max-width:218px; width: 100%;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="715">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="715" align="center" class="lmiczbiodn pt-92921938 pb-93126178 pl-80960861 pr-42946016" style="vertical-align: middle; background-color:#101014; border-radius:12px;  box-shadow: 0px 0px 8px 0px rgba(0, 0, 0, 0.10000000149011612);" bgcolor="#101014">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-90589984 pb-01097341 pl-59594197 pr-99538240" style="vertical-align: middle;   padding-left:64px; padding-right:64px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="64" align="center"><img src="https://media.marka-img.com/983d3508/zFg2NaNL7DLFt9wWt9bQboi7l53M0M.png" width="64" border="0" style="min-width:64px; width:64px;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:24px; height:24px; font-size:24px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<div style="line-height:normal;text-align:center;"><span class="text-34626036 qvhllyhyni" style="color:#ffffff;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:28px;line-height:normal;text-align:center;">Subscription activated successfully!</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-20598375 ykmswwraim" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Hi Krishna,<br><br>Your account has been successfully upgraded to new subscription plan. Thank you for your purchase!<br><br>Here are your plan details:</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-28942147 mltdnzchhf" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Type: </span><span class="text-28942147 cyjptgemve" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Subscription</span></div>
<div style="height:8px;line-height:8px;font-size:8px;">&nbsp;</div>
<div style="line-height:normal;text-align:left;line-height:normal;"><span class="text-28942147 cyjptgemve" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;"></span><span class="text-28942147 wcokmlniwt" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Plan: </span><span class="text-28942147 iaeczfjgoe" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Growth</span></div>
<div style="height:8px;line-height:8px;font-size:8px;">&nbsp;</div>
<div style="line-height:normal;text-align:left;line-height:normal;"><span class="text-28942147 iaeczfjgoe" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;"></span><span class="text-28942147 nlktnbavtq" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Amount: </span><span class="text-28942147 gtblrunjln" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">$34</span></div>
<div style="height:8px;line-height:8px;font-size:8px;">&nbsp;</div>
<div style="line-height:normal;text-align:left;line-height:normal;"><span class="text-28942147 gtblrunjln" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;"></span><span class="text-28942147 grxamfkixj" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">No. of credits: </span><span class="text-28942147 ipgtmngmac" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">7,777</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-58075878 rnsxjemugy" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">You can download your invoice using the link below:</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="right" style="vertical-align: middle; height:45px; background-color:#11e59e; border-radius:8px; border:1.340000033378601px solid #11e59e; box-shadow: 0px 1.340000033378601px 2.680000066757202px 0px rgba(16, 24, 40, 0.05000000074505806);" bgcolor="#11e59e">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle; height:45px;   padding-left:24px; padding-right:24px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;">
<div style="line-height:normal;text-align:center;"><span style="color:#000000;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:18px;line-height:normal;text-align:center;"><a style="color:#000000;text-decoration:none;" href="ReplaceInvoiceURL" target="_blank">Download PDF</a></span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span class="text-30287625 xrjifbgfyh" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:center;">Enjoy your upgraded benefits!</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-02060215 pb-55703351 pl-20314444 pr-93347059 mobile-center" style="vertical-align: middle;  ">
<table border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/privacy-policy">
<span class="text-81430652">Privacy Policy</span></a></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span style="color:#c0c5cd;font-weight:700;font-family:Poppins,Arial,sans-serif;font-size:16px;line-height:normal;text-align:center;">•</span></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/">
<span class="text-93263918">Contact Us</span></a></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.linkedin.com/company/trugen-ai/"><img src="https://media.marka-img.com/983d3508/XW1XmdVyVUWQIGbbs3msSfuyuT2UYU.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
<td style="width:16px; min-width:16px;" width="16">&#8202;</td>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.youtube.com/@trugen_ai"><img src="https://media.marka-img.com/983d3508/UpyDklb7dE1KSmzftPM2yKlKncy6Wz.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span class="text-80601861 bdrheuatzh" style="color:#ffffff;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;">Copyright © Trugen.ai</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</div>
</body>

</html>`
const EMAIL_TEMPLATE_CREDITS = `<!doctype html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">

<head>
<meta charset="utf-8" />
<meta content="width=device-width" name="viewport" />
<meta content="IE=edge" http-equiv="X-UA-Compatible" />
<meta name="x-apple-disable-message-reformatting" />
<meta content="telephone=no,address=no,email=no,date=no,url=no" name="format-detection" />
<title>Credits added to your account</title>
<!--[if mso]>
            <style>
                * {
                    font-family: sans-serif !important;
                }
            </style>
        <![endif]-->
<!--[if !mso]><!-->
<!-- <![endif]-->
<link href="https://fonts.googleapis.com/css?family=Inter:600" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:400" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:700" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Inter:500" rel="stylesheet" type="text/css">
<link href="https://fonts.googleapis.com/css?family=Poppins:700" rel="stylesheet" type="text/css">
<style>
html {
    margin: 0 !important;
    padding: 0 !important;
}

* {
    -ms-text-size-adjust: 100%;
    -webkit-text-size-adjust: 100%;
}
td {
    vertical-align: top;
    mso-table-lspace: 0pt !important;
    mso-table-rspace: 0pt !important;
}
a {
    text-decoration: none;
}
img {
    -ms-interpolation-mode:bicubic;
}
@media only screen and (min-device-width: 320px) and (max-device-width: 374px) {
    u ~ div .email-container {
        min-width: 320px !important;
    }
}
@media only screen and (min-device-width: 375px) and (max-device-width: 413px) {
    u ~ div .email-container {
        min-width: 375px !important;
    }
}
@media only screen and (min-device-width: 414px) {
    u ~ div .email-container {
        min-width: 414px !important;
    }
}

</style>
<!--[if gte mso 9]>
        <xml>
            <o:OfficeDocumentSettings>
                <o:AllowPNG/>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
        <![endif]-->
<style>
@media only screen and (max-device-width: 898px), only screen and (max-width: 898px) {

    .eh {
        height:auto !important;
    }
    .desktop {
        display: none !important;
        height: 0 !important;
        margin: 0 !important;
        max-height: 0 !important;
        overflow: hidden !important;
        padding: 0 !important;
        visibility: hidden !important;
        width: 0 !important;
        min-height: 0 !important;
    }
    .mobile {
        display: block !important;
        width: auto !important;
        height: auto !important;
        float: none !important;
    }
    .email-container {
        width: 100% !important;
        margin: auto !important;
    }
    .stack-column,
    .stack-column-center {
        display: block !important;
        width: 100% !important;
        max-width: 100% !important;
        direction: ltr !important;
    }
    .wid-auto {
        width:auto !important;
    }

    .table-w-full-mobile {
        width: 100%;
    }
    .full-button {
        width: 100% !important;
    }

    .text-31326380 {font-size:20px !important;}.text-92602359 {font-size:16px !important;}.text-00616181 {font-size:16px !important;}.text-23278715 {font-size:16px !important;}.text-86554049 {font-size:16px !important;}.text-06182589 {font-size:12px !important;}.text-40101325 {font-size:12px !important;}.text-36697957 {font-size:12px !important;}.text-42161004 {font-size:14px !important;}
    .pt-69861106 {padding-top:20px !important;}.pb-09550263 {padding-bottom:20px !important;}.pl-53830995 {padding-left:20px !important;}.pr-06539823 {padding-right:20px !important;}.pt-66830480 {padding-top:24px !important;}.pb-97962657 {padding-bottom:24px !important;}.pl-31457102 {padding-left:14px !important;}.pr-25681779 {padding-right:14px !important;}.pt-99099072 {padding-top:16px !important;}.pb-50167104 {padding-bottom:16px !important;}.pl-46428910 {padding-left:16px !important;}.pr-41391232 {padding-right:16px !important;}.pt-66240282 {padding-top:16px !important;}.pb-28784484 {padding-bottom:16px !important;}.pl-95071408 {padding-left:16px !important;}.pr-35108688 {padding-right:16px !important;}

    .mobile-center {
        text-align: center;
    }

    .mobile-center > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-left {
        text-align: left;
    }

    .mobile-left > table {
        display: inline-block;
        vertical-align: inherit;
    }

    .mobile-right {
        text-align: right;
    }

    .mobile-right > table {
        display: inline-block;
        vertical-align: inherit;
    }

}
@media (prefers-color-scheme:dark){ .vkhszkluxg {background-color:#161616 !important} .erlsvqzkno {background-color:#333232 !important} .gdolxmeqet {color: #ffffff !important} .prngslcxvq {color: #BFBFBF !important} .tnvkohwing {color: #BFBFBF !important} .fhvikqynxv {color: #BFBFBF !important} .qnbudsytph {color: #BFBFBF !important} .wmfpytnmct {color: #BFBFBF !important} .pabfforqna {color: #BFBFBF !important} .sdevcivfls {color: #BFBFBF !important} .qupvizcdyy {color: #BFBFBF !important} .lqzkwdyfyx {color: #BFBFBF !important} .ppnigkcqyu {color: #ffffff !important} } 
</style>
</head>

<body width="100%" style="background-color:#dedede;margin:0;padding:0!important;mso-line-height-rule:exactly;">
<div style="background-color:#dedede">
<!--[if gte mso 9]>
                                            <v:background xmlns:v="urn:schemas-microsoft-com:vml" fill="t">
                                            <v:fill type="tile" color="#dedede"/>
                                            </v:background>
                                            <![endif]-->
<table width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td valign="top" align="center">
<table bgcolor="#000000" style="margin:0 auto;" align="center" id="brick_container" cellspacing="0" cellpadding="0" border="0" width="899" class="email-container">
<tr>
<td width="899">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="899" align="center" class="vkhszkluxg pt-69861106 pb-09550263 pl-53830995 pr-06539823" style="background-color:#000000;   padding-left:120px; padding-right:120px;" bgcolor="#000000">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="414" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="218" align="center"><img src="https://media.marka-img.com/983d3508/JrPEXhJRzEILH17JUVzOF4IGwWLHTx.png" width="218" border="0" style="max-width:218px; width: 100%;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="715">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="715" align="center" class="erlsvqzkno pt-66830480 pb-97962657 pl-31457102 pr-25681779" style="vertical-align: middle; background-color:#101014; border-radius:12px;  box-shadow: 0px 0px 8px 0px rgba(0, 0, 0, 0.10000000149011612);" bgcolor="#101014">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-99099072 pb-50167104 pl-46428910 pr-41391232" style="vertical-align: middle;   padding-left:64px; padding-right:64px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="64" align="center"><img src="https://media.marka-img.com/983d3508/ZV6VTCQLSyASQzqdaEtqmOIOQaDrGk.png" width="64" border="0" style="min-width:64px; width:64px;
         height: auto; display: block;"></td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:24px; height:24px; font-size:24px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<div style="line-height:normal;text-align:center;"><span class="text-31326380 gdolxmeqet" style="color:#ffffff;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:28px;line-height:normal;text-align:center;">New credits added to your account!</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-92602359 prngslcxvq" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Hi Krishna,<br><br>Your recent purchase of add-on credits was successful. Thank you for your payment!<br><br>Here are the details of your purchase:</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-00616181 tnvkohwing" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Type: </span><span class="text-00616181 fhvikqynxv" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Add-on Credits </span></div>
<div style="height:8px;line-height:8px;font-size:8px;">&nbsp;</div>
<div style="line-height:normal;text-align:left;line-height:normal;"><span class="text-00616181 fhvikqynxv" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;"></span><span class="text-00616181 qnbudsytph" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">Amount: </span><span class="text-00616181 wmfpytnmct" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">$34</span></div>
<div style="height:8px;line-height:8px;font-size:8px;">&nbsp;</div>
<div style="line-height:normal;text-align:left;line-height:normal;"><span class="text-00616181 wmfpytnmct" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;"></span><span class="text-00616181 pabfforqna" style="color:#d0d0d0;font-weight:700;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">No. of Additional Credits: </span><span class="text-00616181 sdevcivfls" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">10,000</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:left;"><span class="text-23278715 qupvizcdyy" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:left;">You can download your invoice using the link below:</span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="right" style="vertical-align: middle; height:45px; background-color:#11e59e; border-radius:8px; border:1.340000033378601px solid #11e59e; box-shadow: 0px 1.340000033378601px 2.680000066757202px 0px rgba(16, 24, 40, 0.05000000074505806);" bgcolor="#11e59e">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle; height:45px;   padding-left:24px; padding-right:24px;">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;">
<div style="line-height:normal;text-align:center;"><span style="color:#000000;font-weight:600;font-family:Inter,Arial,sans-serif;font-size:18px;line-height:normal;text-align:center;"><a style="color:#000000;text-decoration:none;" href="ReplaceInvoiceURL" target="_blank">Download PDF</a></span></div>
</td>
</tr>
<tr>
<td>
<div style="line-height:16px; height:16px; font-size:16px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:38px; height:38px; font-size:38px">&#8202;</div>
</td>
</tr>
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span class="text-86554049 lqzkwdyfyx" style="color:#d0d0d0;font-family:Inter,Arial,sans-serif;font-size:20px;line-height:normal;text-align:center;">Your new credits are now available for use. Enjoy!</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:32px; height:32px; font-size:32px">&#8202;</div>
</td>
</tr>
<tr>
<td width="100%" align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="100%">
<table width="100%" cellspacing="0" cellpadding="0" border="0">
<tr>
<td width="100%" align="center" class="pt-66240282 pb-28784484 pl-95071408 pr-35108688 mobile-center" style="vertical-align: middle;  ">
<table border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/privacy-policy">
<span class="text-06182589">Privacy Policy</span></a></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span style="color:#c0c5cd;font-weight:700;font-family:Poppins,Arial,sans-serif;font-size:16px;line-height:normal;text-align:center;">•</span></div>
</td>
<td style="width:4px; min-width:4px;" width="4">&#8202;</td>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><a style="color:#11e59e;font-weight:500;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;text-decoration:underline;" href="https://trugen.ai/">
<span class="text-36697957">Contact Us</span></a></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td align="center" style="vertical-align: middle;  ">
<table width="100%" border="0" cellpadding="0" cellspacing="0">
<tr>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.linkedin.com/company/trugen-ai/"><img src="https://media.marka-img.com/983d3508/iCoc7royp4RpIL7h0E9ELmAvAh42gR.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
<td style="width:16px; min-width:16px;" width="16">&#8202;</td>
<td style="vertical-align: middle;" width="48" align="center"><a href="https://www.youtube.com/@trugen_ai"><img src="https://media.marka-img.com/983d3508/dj3gHLosq81QaNZNpWlEYNbIK5qnzx.png" width="48" border="0" style="min-width:48px; width:48px;
        border-radius:6px; height: auto; display: block;"></a></td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<div style="line-height:20px; height:20px; font-size:20px">&#8202;</div>
</td>
</tr>
<tr>
<td class="mobile-center" align="center">
<table cellspacing="0" cellpadding="0" border="0">
<tr>
<td style="vertical-align: middle;" align="center">
<div style="line-height:normal;text-align:center;"><span class="text-42161004 ppnigkcqyu" style="color:#ffffff;font-family:Inter,Arial,sans-serif;font-size:14px;line-height:normal;text-align:center;">Copyright © Trugen.ai</span></div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
<tr class="desktop">
<td>
<div style="line-height:64px; height:64px; font-size:64px">&#8202;</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</div>
</body>

</html>`

func GetEmailTemplateHTML(templateName string) string {
	switch templateName {
	case "invitation":
		return EMAIL_TEMPLATE_INVITATION
	case "subscription":
		return EMAIL_TEMPLATE_SUBSCRIPTION
	case "credits":
		return EMAIL_TEMPLATE_CREDITS
	default:
		return ""
	}
}

func GetAccessToken(tenantID, clientID, clientSecret string) (string, error) {
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
	data := []byte(fmt.Sprintf(
		"client_id=%s&scope=https%%3A%%2F%%2Fgraph.microsoft.com%%2F.default&client_secret=%s&grant_type=client_credentials",
		clientID, clientSecret))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token: %s", body)
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// SendEmail sends an HTML email using Azure OAuth email API.
func SendEmail(to, subject, htmlBody string) error {
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	fromEmail := os.Getenv("AZURE_FROM_EMAIL")

	// Validate env config
	if tenantID == "" || clientID == "" || clientSecret == "" || fromEmail == "" {
		return fmt.Errorf("missing Azure email configuration")
	}

	// Get OAuth token
	token, err := GetAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("token error: %w", err)
	}

	// Send email
	if err := SendHTMLEmail(token, fromEmail, to, subject, htmlBody); err != nil {
		return fmt.Errorf("email send error: %w", err)
	}

	return nil
}

func SendHTMLEmail(accessToken, from, to, subject, htmlBody string) error {
	url := "https://graph.microsoft.com/v1.0/users/" + from + "/sendMail"

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": "HTML",
				"content":     htmlBody,
			},
			"toRecipients": []map[string]map[string]string{
				{"emailAddress": {"address": to}},
			},
		},
		"saveToSentItems": "true",
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("email send failed: %s", body)
	}

	fmt.Println("✅ Email sent successfully!")
	return nil
}
