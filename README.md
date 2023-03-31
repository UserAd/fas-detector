# Simple FAS detector for AsteriskPBX
FAS (False Answer Supervisor) is a technique many VoIP providers use to extend the call duration (and cost) of a call.

FAS can be the following types:

1. Answer and send silence or progress tones to the caller
2. Adding silence or playing back the recording of the call after callee hangups
3. Rounding (or adding) duration on the billing side of the operator.

Variation of sending ring tone after an answer is easily detectable but still used by some operators.

## Ring tones around the world
All ring tones created by transmitting well-known frequencies:

1. 425Hz
2. 440Hz (in France)
3. 450Hz (in China)
4. 480Hz (one of Bell’s tones)
5. 400Hz (one of the UK/Ireland tones)

For detecting these tones we can get a record of the call, split it to windows and check frequencies with FFT. If some of these frequencies are dominant in a window then we can say that we hear a ringtone.

## Dialplan example

```
[fas]
exten => _XX.,1,Set(filename=${EXTEN}-${EPOCH})
exten => _XX.,n,MixMonitor(/tmp/${filename}.wav,b)
exten => _XX.,n,Dial(SIP/trunk/${EXTEN})

exten => h,1,GotoIf($["${DIALSTATUS}" = "ANSWER"]?detect:exit)

exten => h,n(detect),StopMixMonitor()
exten => h,n,Agi(/usr/local/bin/fas-detector,${filename})
exten => h,n,Set(CDR(fas_detected)=${FAS_DETECTED})
exten => h,n(exit),NoOp

```

In this example, I use the MixMonitor application to record call after answer, execute the detector application via an AGI interface and write the result in the CDR field named fas\_detected.

## FAS detector application

In the source code that you can find in this GitHub repository, you can see what we read samples from a file, skip the first second, check the next five seconds and write the result as AGI call to variable FAS\_DETECTED.

## Conclusion
Unfortunately, this method can have false-negative and false-positive events. For example, if the call is connected to a voicemail box. And can’t be used on an operator scale.

For more accurate results detection can be performed by deep learning algorithms with some statistical analysis of other calls. But this filter can be used for collecting data for learning other models.
