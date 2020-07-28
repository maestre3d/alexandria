# Trending Algorithm

It is essential to consider specific factors in order to get the best potential result for trending detection. 
The first thing to consider is the _increment or decrement ratio between two y-axis in a given time (x-axis)_ to detect a peak, you may use deltas _(Δf(x))_ 
for this. 
You would have the following mathematical model:


_**ω** = (Δf(x)) / f(x) or **ω** = (Δy) / y_


Where **_y_** or **_f(x)_** represents _social interaction scoring (SI)_ while **_ω_** represents the _**Social Fluctuation Ratio** (SFR)_. Furthermore, SI represents a _view, like, mention or comment_ in an specific aggregate. However, by using this specific ratio it is impossible to detect peaks in a _high-mass 
environment_ such as the internet that is an Exponential-Driven and high-mass environment. SFR or ω is highly relative when comparison are being made in a high-mass environment.


Consider the following potential scenario: _aggregate "x"_ has 1 social interaction (SI) at the first hour (t), eventually it went from 1 to 10 SI in 
the next hour. Furthermore, _aggregate "y"_ has an increment of 123,980 to 125,980 SI in 1 hour (t). Henceforward, you may calculate the increment or 
decrement ratio (social fluctuation ratio or ω) to determine if there was a peak between the given hours (t). 

Using the past hypothetical scenario, 
aggregate "x" social fluctuation ratio (ω) is **0.9 (90%)** while aggregate "y" ω is **0.1587 (1.87%)**. Consequently, if you exclusively used this 
mathematical model, _you won't get the result needed to calculate trends as discussed before._


Thus, you may consider **two factors** as key to discriminate each candidate. While ω (social fluctuation ratio) works to get the SI increment and 
decrement ratio, it lacks of a required factor to _remove the high-relativity_. Here is when the **_Valuable Ratio (υ)_** is mentioned. Valuable ratio 
mathematical model is represented like this:

_**υ** = f(x)/z_ or _**υ** = y/z_

Where **f(x)** represents the given _SI (social interaction score)_ and **z** represents the _maximum SI score_ in the whole data set.

This **valuable ratio (υ)** gives you the _required factor to eradicate high-relativity_ in a large data set. Therefore, you may be able to calculate 
the required score to detect a trend in a Exponential-Driven and high-mass environment.

To conclude, _you must sum the given mathematical algorithms_ to get an overall score called **Trending Score** _(represented as Φ)_. It may be represented 
in a mathematical model as this:


_**Φ** = ω + υ_


Where **υ** represents _valuable ratio_, **ω** represents _social fluctuation ratio_ and **Φ** represents _trending score_. In addition, this lends us to _store these results in a specialized database_ such as a data lake to query any aggregate by it's high-cardinality field and output very specific results for each 
scenario _(like trends for an specific user based on his consumed content and country by denormalizing aggregate fields)_.

Based on the past mathematical model, you may calculate the Trending Score using the past hypothetical scenario.


- _Aggregate "x"_ Trending Score = Φ(1)
- _Aggregate "y"_ Trending Score = Φ(2)
- _Maximum SI score_ = 150,000


Φ(1) = ((10-1)/10) + (10/150,000) = 0.900066667


Φ(2) = ((125,980-123,980)/125,980) + (125,980/150,000) = 0.855742202
