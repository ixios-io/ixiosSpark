/* bignumber.js v2.0.3 https://github.com/MikeMcl/bignumber.js/LICENCE */
/* modified by zelig to fix https://github.com/robertkrimen/otto#regular-expression-incompatibility */
!function (e) {
    "use strict";

    function n(e) {
        function a(e, n) {
            var t, r, i, o, u, s, f = this;
            if (!(f instanceof a)) return j && L(26, "constructor call without new", e), new a(e, n);
            if (null != n && H(n, 2, 64, M, "base")) {
                if (n = 0 | n, s = e + "", 10 == n) return f = new a(e instanceof a ? e : s), U(f, P + f.e + 1, k);
                if ((o = "number" == typeof e) && 0 * e != 0 || !new RegExp("^-?" + (t = "[" + O.slice(0, n) + "]+") + "(?:\\." + t + ")?$", 37 > n ? "i" : "").test(s)) return g(f, s, o, n);
                o ? (f.s = 0 > 1 / e ? (s = s.slice(1), -1) : 1, j && s.replace(/^0\.0*|\./, "").length > 15 && L(M, b, e), o = !1) : f.s = 45 === s.charCodeAt(0) ? (s = s.slice(1), -1) : 1, s = D(s, 10, n, f.s)
            } else {
                if (e instanceof a) return f.s = e.s, f.e = e.e, f.c = (e = e.c) ? e.slice() : e, void (M = 0);
                if ((o = "number" == typeof e) && 0 * e == 0) {
                    if (f.s = 0 > 1 / e ? (e = -e, -1) : 1, e === ~~e) {
                        for (r = 0, i = e; i >= 10; i /= 10, r++) ;
                        return f.e = r, f.c = [e], void (M = 0)
                    }
                    s = e + ""
                } else {
                    if (!p.test(s = e + "")) return g(f, s, o);
                    f.s = 45 === s.charCodeAt(0) ? (s = s.slice(1), -1) : 1
                }
            }
            for ((r = s.indexOf(".")) > -1 && (s = s.replace(".", "")), (i = s.search(/e/i)) > 0 ? (0 > r && (r = i), r += +s.slice(i + 1), s = s.substring(0, i)) : 0 > r && (r = s.length), i = 0; 48 === s.charCodeAt(i); i++) ;
            for (u = s.length; 48 === s.charCodeAt(--u);) ;
            if (s = s.slice(i, u + 1)) if (u = s.length, o && j && u > 15 && L(M, b, f.s * e), r = r - i - 1, r > z) f.c = f.e = null; else if (G > r) f.c = [f.e = 0]; else {
                if (f.e = r, f.c = [], i = (r + 1) % y, 0 > r && (i += y), u > i) {
                    for (i && f.c.push(+s.slice(0, i)), u -= y; u > i;) f.c.push(+s.slice(i, i += y));
                    s = s.slice(i), i = y - s.length
                } else i -= u;
                for (; i--; s += "0") ;
                f.c.push(+s)
            } else f.c = [f.e = 0];
            M = 0
        }

        function D(e, n, t, i) {
            var o, u, f, c, h, g, p, d = e.indexOf("."), m = P, w = k;
            for (37 > t && (e = e.toLowerCase()), d >= 0 && (f = J, J = 0, e = e.replace(".", ""), p = new a(t), h = p.pow(e.length - d), J = f, p.c = s(l(r(h.c), h.e), 10, n), p.e = p.c.length), g = s(e, t, n), u = f = g.length; 0 == g[--f]; g.pop()) ;
            if (!g[0]) return "0";
            if (0 > d ? --u : (h.c = g, h.e = u, h.s = i, h = C(h, p, m, w, n), g = h.c, c = h.r, u = h.e), o = u + m + 1, d = g[o], f = n / 2, c = c || 0 > o || null != g[o + 1], c = 4 > w ? (null != d || c) && (0 == w || w == (h.s < 0 ? 3 : 2)) : d > f || d == f && (4 == w || c || 6 == w && 1 & g[o - 1] || w == (h.s < 0 ? 8 : 7)), 1 > o || !g[0]) e = c ? l("1", -m) : "0"; else {
                if (g.length = o, c) for (--n; ++g[--o] > n;) g[o] = 0, o || (++u, g.unshift(1));
                for (f = g.length; !g[--f];) ;
                for (d = 0, e = ""; f >= d; e += O.charAt(g[d++])) ;
                e = l(e, u)
            }
            return e
        }

        function _(e, n, t, i) {
            var o, u, s, c, h;
            if (t = null != t && H(t, 0, 8, i, v) ? 0 | t : k, !e.c) return e.toString();
            if (o = e.c[0], s = e.e, null == n) h = r(e.c), h = 19 == i || 24 == i && B >= s ? f(h, s) : l(h, s); else if (e = U(new a(e), n, t), u = e.e, h = r(e.c), c = h.length, 19 == i || 24 == i && (u >= n || B >= u)) {
                for (; n > c; h += "0", c++) ;
                h = f(h, u)
            } else if (n -= s, h = l(h, u), u + 1 > c) {
                if (--n > 0) for (h += "."; n--; h += "0") ;
            } else if (n += u - c, n > 0) for (u + 1 == c && (h += "."); n--; h += "0") ;
            return e.s < 0 && o ? "-" + h : h
        }

        function x(e, n) {
            var t, r, i = 0;
            for (u(e[0]) && (e = e[0]), t = new a(e[0]); ++i < e.length;) {
                if (r = new a(e[i]), !r.s) {
                    t = r;
                    break
                }
                n.call(t, r) && (t = r)
            }
            return t
        }

        function F(e, n, t, r, i) {
            return (n > e || e > t || e != c(e)) && L(r, (i || "decimal places") + (n > e || e > t ? " out of range" : " not an integer"), e), !0
        }

        function I(e, n, t) {
            for (var r = 1, i = n.length; !n[--i]; n.pop()) ;
            for (i = n[0]; i >= 10; i /= 10, r++) ;
            return (t = r + t * y - 1) > z ? e.c = e.e = null : G > t ? e.c = [e.e = 0] : (e.e = t, e.c = n), e
        }

        function L(e, n, t) {
            var r = new Error(["new BigNumber", "cmp", "config", "div", "divToInt", "eq", "gt", "gte", "lt", "lte", "minus", "mod", "plus", "precision", "random", "round", "shift", "times", "toDigits", "toExponential", "toFixed", "toFormat", "toFraction", "pow", "toPrecision", "toString", "BigNumber"][e] + "() " + n + ": " + t);
            throw r.name = "BigNumber Error", M = 0, r
        }

        function U(e, n, t, r) {
            var i, o, u, s, f, l, c, a = e.c, h = R;
            if (a) {
                e:{
                    for (i = 1, s = a[0]; s >= 10; s /= 10, i++) ;
                    if (o = n - i, 0 > o) o += y, u = n, f = a[l = 0], c = f / h[i - u - 1] % 10 | 0; else if (l = d((o + 1) / y), l >= a.length) {
                        if (!r) break e;
                        for (; a.length <= l; a.push(0)) ;
                        f = c = 0, i = 1, o %= y, u = o - y + 1
                    } else {
                        for (f = s = a[l], i = 1; s >= 10; s /= 10, i++) ;
                        o %= y, u = o - y + i, c = 0 > u ? 0 : f / h[i - u - 1] % 10 | 0
                    }
                    if (r = r || 0 > n || null != a[l + 1] || (0 > u ? f : f % h[i - u - 1]), r = 4 > t ? (c || r) && (0 == t || t == (e.s < 0 ? 3 : 2)) : c > 5 || 5 == c && (4 == t || r || 6 == t && (o > 0 ? u > 0 ? f / h[i - u] : 0 : a[l - 1]) % 10 & 1 || t == (e.s < 0 ? 8 : 7)), 1 > n || !a[0]) return a.length = 0, r ? (n -= e.e + 1, a[0] = h[n % y], e.e = -n || 0) : a[0] = e.e = 0, e;
                    if (0 == o ? (a.length = l, s = 1, l--) : (a.length = l + 1, s = h[y - o], a[l] = u > 0 ? m(f / h[i - u] % h[u]) * s : 0), r) for (; ;) {
                        if (0 == l) {
                            for (o = 1, u = a[0]; u >= 10; u /= 10, o++) ;
                            for (u = a[0] += s, s = 1; u >= 10; u /= 10, s++) ;
                            o != s && (e.e++, a[0] == N && (a[0] = 1));
                            break
                        }
                        if (a[l] += s, a[l] != N) break;
                        a[l--] = 0, s = 1
                    }
                    for (o = a.length; 0 === a[--o]; a.pop()) ;
                }
                e.e > z ? e.c = e.e = null : e.e < G && (e.c = [e.e = 0])
            }
            return e
        }

        var C, M = 0, T = a.prototype, q = new a(1), P = 20, k = 4, B = -7, $ = 21, G = -1e7, z = 1e7, j = !0, H = F,
            V = !1, W = 1, J = 100, X = {
                decimalSeparator: ".",
                groupSeparator: ",",
                groupSize: 3,
                secondaryGroupSize: 0,
                fractionGroupSeparator: " ",
                fractionGroupSize: 0
            };
        return a.another = n, a.ROUND_UP = 0, a.ROUND_DOWN = 1, a.ROUND_CEIL = 2, a.ROUND_FLOOR = 3, a.ROUND_HALF_UP = 4, a.ROUND_HALF_DOWN = 5, a.ROUND_HALF_EVEN = 6, a.ROUND_HALF_CEIL = 7, a.ROUND_HALF_FLOOR = 8, a.EUCLID = 9, a.config = function () {
            var e, n, t = 0, r = {}, i = arguments, s = i[0], f = s && "object" == typeof s ? function () {
                return s.hasOwnProperty(n) ? null != (e = s[n]) : void 0
            } : function () {
                return i.length > t ? null != (e = i[t++]) : void 0
            };
            return f(n = "DECIMAL_PLACES") && H(e, 0, E, 2, n) && (P = 0 | e), r[n] = P, f(n = "ROUNDING_MODE") && H(e, 0, 8, 2, n) && (k = 0 | e), r[n] = k, f(n = "EXPONENTIAL_AT") && (u(e) ? H(e[0], -E, 0, 2, n) && H(e[1], 0, E, 2, n) && (B = 0 | e[0], $ = 0 | e[1]) : H(e, -E, E, 2, n) && (B = -($ = 0 | (0 > e ? -e : e)))), r[n] = [B, $], f(n = "RANGE") && (u(e) ? H(e[0], -E, -1, 2, n) && H(e[1], 1, E, 2, n) && (G = 0 | e[0], z = 0 | e[1]) : H(e, -E, E, 2, n) && (0 | e ? G = -(z = 0 | (0 > e ? -e : e)) : j && L(2, n + " cannot be zero", e))), r[n] = [G, z], f(n = "ERRORS") && (e === !!e || 1 === e || 0 === e ? (M = 0, H = (j = !!e) ? F : o) : j && L(2, n + w, e)), r[n] = j, f(n = "CRYPTO") && (e === !!e || 1 === e || 0 === e ? (V = !(!e || !h || "object" != typeof h), e && !V && j && L(2, "crypto unavailable", h)) : j && L(2, n + w, e)), r[n] = V, f(n = "MODULO_MODE") && H(e, 0, 9, 2, n) && (W = 0 | e), r[n] = W, f(n = "POW_PRECISION") && H(e, 0, E, 2, n) && (J = 0 | e), r[n] = J, f(n = "FORMAT") && ("object" == typeof e ? X = e : j && L(2, n + " not an object", e)), r[n] = X, r
        }, a.max = function () {
            return x(arguments, T.lt)
        }, a.min = function () {
            return x(arguments, T.gt)
        }, a.random = function () {
            var e = 9007199254740992, n = Math.random() * e & 2097151 ? function () {
                return m(Math.random() * e)
            } : function () {
                return 8388608 * (1073741824 * Math.random() | 0) + (8388608 * Math.random() | 0)
            };
            return function (e) {
                var t, r, i, o, u, s = 0, f = [], l = new a(q);
                if (e = null != e && H(e, 0, E, 14) ? 0 | e : P, o = d(e / y), V) if (h && h.getRandomValues) {
                    for (t = h.getRandomValues(new Uint32Array(o *= 2)); o > s;) u = 131072 * t[s] + (t[s + 1] >>> 11), u >= 9e15 ? (r = h.getRandomValues(new Uint32Array(2)), t[s] = r[0], t[s + 1] = r[1]) : (f.push(u % 1e14), s += 2);
                    s = o / 2
                } else if (h && h.randomBytes) {
                    for (t = h.randomBytes(o *= 7); o > s;) u = 281474976710656 * (31 & t[s]) + 1099511627776 * t[s + 1] + 4294967296 * t[s + 2] + 16777216 * t[s + 3] + (t[s + 4] << 16) + (t[s + 5] << 8) + t[s + 6], u >= 9e15 ? h.randomBytes(7).copy(t, s) : (f.push(u % 1e14), s += 7);
                    s = o / 7
                } else j && L(14, "crypto unavailable", h);
                if (!s) for (; o > s;) u = n(), 9e15 > u && (f[s++] = u % 1e14);
                for (o = f[--s], e %= y, o && e && (u = R[y - e], f[s] = m(o / u) * u); 0 === f[s]; f.pop(), s--) ;
                if (0 > s) f = [i = 0]; else {
                    for (i = -1; 0 === f[0]; f.shift(), i -= y) ;
                    for (s = 1, u = f[0]; u >= 10; u /= 10, s++) ;
                    y > s && (i -= y - s)
                }
                return l.e = i, l.c = f, l
            }
        }(), C = function () {
            function e(e, n, t) {
                var r, i, o, u, s = 0, f = e.length, l = n % A, c = n / A | 0;
                for (e = e.slice(); f--;) o = e[f] % A, u = e[f] / A | 0, r = c * o + u * l, i = l * o + r % A * A + s, s = (i / t | 0) + (r / A | 0) + c * u, e[f] = i % t;
                return s && e.unshift(s), e
            }

            function n(e, n, t, r) {
                var i, o;
                if (t != r) o = t > r ? 1 : -1; else for (i = o = 0; t > i; i++) if (e[i] != n[i]) {
                    o = e[i] > n[i] ? 1 : -1;
                    break
                }
                return o
            }

            function r(e, n, t, r) {
                for (var i = 0; t--;) e[t] -= i, i = e[t] < n[t] ? 1 : 0, e[t] = i * r + e[t] - n[t];
                for (; !e[0] && e.length > 1; e.shift()) ;
            }

            return function (i, o, u, s, f) {
                var l, c, h, g, p, d, w, v, b, O, S, R, A, E, D, _, x, F = i.s == o.s ? 1 : -1, I = i.c, L = o.c;
                if (!(I && I[0] && L && L[0])) return new a(i.s && o.s && (I ? !L || I[0] != L[0] : L) ? I && 0 == I[0] || !L ? 0 * F : F / 0 : 0 / 0);
                for (v = new a(F), b = v.c = [], c = i.e - o.e, F = u + c + 1, f || (f = N, c = t(i.e / y) - t(o.e / y), F = F / y | 0), h = 0; L[h] == (I[h] || 0); h++) ;
                if (L[h] > (I[h] || 0) && c--, 0 > F) b.push(1), g = !0; else {
                    for (E = I.length, _ = L.length, h = 0, F += 2, p = m(f / (L[0] + 1)), p > 1 && (L = e(L, p, f), I = e(I, p, f), _ = L.length, E = I.length), A = _, O = I.slice(0, _), S = O.length; _ > S; O[S++] = 0) ;
                    x = L.slice(), x.unshift(0), D = L[0], L[1] >= f / 2 && D++;
                    do p = 0, l = n(L, O, _, S), 0 > l ? (R = O[0], _ != S && (R = R * f + (O[1] || 0)), p = m(R / D), p > 1 ? (p >= f && (p = f - 1), d = e(L, p, f), w = d.length, S = O.length, l = n(d, O, w, S), 1 == l && (p--, r(d, w > _ ? x : L, w, f))) : (0 == p && (l = p = 1), d = L.slice()), w = d.length, S > w && d.unshift(0), r(O, d, S, f), -1 == l && (S = O.length, l = n(L, O, _, S), 1 > l && (p++, r(O, S > _ ? x : L, S, f))), S = O.length) : 0 === l && (p++, O = [0]), b[h++] = p, l && O[0] ? O[S++] = I[A] || 0 : (O = [I[A]], S = 1); while ((A++ < E || null != O[0]) && F--);
                    g = null != O[0], b[0] || b.shift()
                }
                if (f == N) {
                    for (h = 1, F = b[0]; F >= 10; F /= 10, h++) ;
                    U(v, u + (v.e = h + c * y - 1) + 1, s, g)
                } else v.e = c, v.r = +g;
                return v
            }
        }(), g = function () {
            var e = /^(-?)0([xbo])(\w[\w.]*$)/i, n = /^([^.]+)\.$/, t = /^\.([^.]+)$/, r = /^-?(Infinity|NaN)$/,
                i = /^\s*\+([\w.])|^\s+|\s+$/g;
            return function (o, u, s, f) {
                var l, c = s ? u : u.replace(i, "$1");
                if (r.test(c)) o.s = isNaN(c) ? null : 0 > c ? -1 : 1; else {
                    if (!s && (c = c.replace(e, function (e, n, t) {
                        return l = "x" == (t = t.toLowerCase()) ? 16 : "b" == t ? 2 : 8, f && f != l ? e : n
                    }), f && (l = f, c = c.replace(n, "$1").replace(t, "0.$1")), u != c)) return new a(c, l);
                    j && L(M, "not a" + (f ? " base " + f : "") + " number", u), o.s = null
                }
                o.c = o.e = null, M = 0
            }
        }(), T.absoluteValue = T.abs = function () {
            var e = new a(this);
            return e.s < 0 && (e.s = 1), e
        }, T.ceil = function () {
            return U(new a(this), this.e + 1, 2)
        }, T.comparedTo = T.cmp = function (e, n) {
            return M = 1, i(this, new a(e, n))
        }, T.decimalPlaces = T.dp = function () {
            var e, n, r = this.c;
            if (!r) return null;
            if (e = ((n = r.length - 1) - t(this.e / y)) * y, n = r[n]) for (; n % 10 == 0; n /= 10, e--) ;
            return 0 > e && (e = 0), e
        }, T.dividedBy = T.div = function (e, n) {
            return M = 3, C(this, new a(e, n), P, k)
        }, T.dividedToIntegerBy = T.divToInt = function (e, n) {
            return M = 4, C(this, new a(e, n), 0, 1)
        }, T.equals = T.eq = function (e, n) {
            return M = 5, 0 === i(this, new a(e, n))
        }, T.floor = function () {
            return U(new a(this), this.e + 1, 3)
        }, T.greaterThan = T.gt = function (e, n) {
            return M = 6, i(this, new a(e, n)) > 0
        }, T.greaterThanOrEqualTo = T.gte = function (e, n) {
            return M = 7, 1 === (n = i(this, new a(e, n))) || 0 === n
        }, T.isFinite = function () {
            return !!this.c
        }, T.isInteger = T.isInt = function () {
            return !!this.c && t(this.e / y) > this.c.length - 2
        }, T.isNaN = function () {
            return !this.s
        }, T.isNegative = T.isNeg = function () {
            return this.s < 0
        }, T.isZero = function () {
            return !!this.c && 0 == this.c[0]
        }, T.lessThan = T.lt = function (e, n) {
            return M = 8, i(this, new a(e, n)) < 0
        }, T.lessThanOrEqualTo = T.lte = function (e, n) {
            return M = 9, -1 === (n = i(this, new a(e, n))) || 0 === n
        }, T.minus = T.sub = function (e, n) {
            var r, i, o, u, s = this, f = s.s;
            if (M = 10, e = new a(e, n), n = e.s, !f || !n) return new a(0 / 0);
            if (f != n) return e.s = -n, s.plus(e);
            var l = s.e / y, c = e.e / y, h = s.c, g = e.c;
            if (!l || !c) {
                if (!h || !g) return h ? (e.s = -n, e) : new a(g ? s : 0 / 0);
                if (!h[0] || !g[0]) return g[0] ? (e.s = -n, e) : new a(h[0] ? s : 3 == k ? -0 : 0)
            }
            if (l = t(l), c = t(c), h = h.slice(), f = l - c) {
                for ((u = 0 > f) ? (f = -f, o = h) : (c = l, o = g), o.reverse(), n = f; n--; o.push(0)) ;
                o.reverse()
            } else for (i = (u = (f = h.length) < (n = g.length)) ? f : n, f = n = 0; i > n; n++) if (h[n] != g[n]) {
                u = h[n] < g[n];
                break
            }
            if (u && (o = h, h = g, g = o, e.s = -e.s), n = (i = g.length) - (r = h.length), n > 0) for (; n--; h[r++] = 0) ;
            for (n = N - 1; i > f;) {
                if (h[--i] < g[i]) {
                    for (r = i; r && !h[--r]; h[r] = n) ;
                    --h[r], h[i] += N
                }
                h[i] -= g[i]
            }
            for (; 0 == h[0]; h.shift(), --c) ;
            return h[0] ? I(e, h, c) : (e.s = 3 == k ? -1 : 1, e.c = [e.e = 0], e)
        }, T.modulo = T.mod = function (e, n) {
            var t, r, i = this;
            return M = 11, e = new a(e, n), !i.c || !e.s || e.c && !e.c[0] ? new a(0 / 0) : !e.c || i.c && !i.c[0] ? new a(i) : (9 == W ? (r = e.s, e.s = 1, t = C(i, e, 0, 3), e.s = r, t.s *= r) : t = C(i, e, 0, W), i.minus(t.times(e)))
        }, T.negated = T.neg = function () {
            var e = new a(this);
            return e.s = -e.s || null, e
        }, T.plus = T.add = function (e, n) {
            var r, i = this, o = i.s;
            if (M = 12, e = new a(e, n), n = e.s, !o || !n) return new a(0 / 0);
            if (o != n) return e.s = -n, i.minus(e);
            var u = i.e / y, s = e.e / y, f = i.c, l = e.c;
            if (!u || !s) {
                if (!f || !l) return new a(o / 0);
                if (!f[0] || !l[0]) return l[0] ? e : new a(f[0] ? i : 0 * o)
            }
            if (u = t(u), s = t(s), f = f.slice(), o = u - s) {
                for (o > 0 ? (s = u, r = l) : (o = -o, r = f), r.reverse(); o--; r.push(0)) ;
                r.reverse()
            }
            for (o = f.length, n = l.length, 0 > o - n && (r = l, l = f, f = r, n = o), o = 0; n;) o = (f[--n] = f[n] + l[n] + o) / N | 0, f[n] %= N;
            return o && (f.unshift(o), ++s), I(e, f, s)
        }, T.precision = T.sd = function (e) {
            var n, t, r = this, i = r.c;
            if (null != e && e !== !!e && 1 !== e && 0 !== e && (j && L(13, "argument" + w, e), e != !!e && (e = null)), !i) return null;
            if (t = i.length - 1, n = t * y + 1, t = i[t]) {
                for (; t % 10 == 0; t /= 10, n--) ;
                for (t = i[0]; t >= 10; t /= 10, n++) ;
            }
            return e && r.e + 1 > n && (n = r.e + 1), n
        }, T.round = function (e, n) {
            var t = new a(this);
            return (null == e || H(e, 0, E, 15)) && U(t, ~~e + this.e + 1, null != n && H(n, 0, 8, 15, v) ? 0 | n : k), t
        }, T.shift = function (e) {
            var n = this;
            return H(e, -S, S, 16, "argument") ? n.times("1e" + c(e)) : new a(n.c && n.c[0] && (-S > e || e > S) ? n.s * (0 > e ? 0 : 1 / 0) : n)
        }, T.squareRoot = T.sqrt = function () {
            var e, n, i, o, u, s = this, f = s.c, l = s.s, c = s.e, h = P + 4, g = new a("0.5");
            if (1 !== l || !f || !f[0]) return new a(!l || 0 > l && (!f || f[0]) ? 0 / 0 : f ? s : 1 / 0);
            if (l = Math.sqrt(+s), 0 == l || l == 1 / 0 ? (n = r(f), (n.length + c) % 2 == 0 && (n += "0"), l = Math.sqrt(n), c = t((c + 1) / 2) - (0 > c || c % 2), l == 1 / 0 ? n = "1e" + c : (n = l.toExponential(), n = n.slice(0, n.indexOf("e") + 1) + c), i = new a(n)) : i = new a(l + ""), i.c[0]) for (c = i.e, l = c + h, 3 > l && (l = 0); ;) if (u = i, i = g.times(u.plus(C(s, u, h, 1))), r(u.c).slice(0, l) === (n = r(i.c)).slice(0, l)) {
                if (i.e < c && --l, n = n.slice(l - 3, l + 1), "9999" != n && (o || "4999" != n)) {
                    (!+n || !+n.slice(1) && "5" == n.charAt(0)) && (U(i, i.e + P + 2, 1), e = !i.times(i).eq(s));
                    break
                }
                if (!o && (U(u, u.e + P + 2, 0), u.times(u).eq(s))) {
                    i = u;
                    break
                }
                h += 4, l += 4, o = 1
            }
            return U(i, i.e + P + 1, k, e)
        }, T.times = T.mul = function (e, n) {
            var r, i, o, u, s, f, l, c, h, g, p, d, m, w, v, b = this, O = b.c, S = (M = 17, e = new a(e, n)).c;
            if (!(O && S && O[0] && S[0])) return !b.s || !e.s || O && !O[0] && !S || S && !S[0] && !O ? e.c = e.e = e.s = null : (e.s *= b.s, O && S ? (e.c = [0], e.e = 0) : e.c = e.e = null), e;
            for (i = t(b.e / y) + t(e.e / y), e.s *= b.s, l = O.length, g = S.length, g > l && (m = O, O = S, S = m, o = l, l = g, g = o), o = l + g, m = []; o--; m.push(0)) ;
            for (w = N, v = A, o = g; --o >= 0;) {
                for (r = 0, p = S[o] % v, d = S[o] / v | 0, s = l, u = o + s; u > o;) c = O[--s] % v, h = O[s] / v | 0, f = d * c + h * p, c = p * c + f % v * v + m[u] + r, r = (c / w | 0) + (f / v | 0) + d * h, m[u--] = c % w;
                m[u] = r
            }
            return r ? ++i : m.shift(), I(e, m, i)
        }, T.toDigits = function (e, n) {
            var t = new a(this);
            return e = null != e && H(e, 1, E, 18, "precision") ? 0 | e : null, n = null != n && H(n, 0, 8, 18, v) ? 0 | n : k, e ? U(t, e, n) : t
        }, T.toExponential = function (e, n) {
            return _(this, null != e && H(e, 0, E, 19) ? ~~e + 1 : null, n, 19)
        }, T.toFixed = function (e, n) {
            return _(this, null != e && H(e, 0, E, 20) ? ~~e + this.e + 1 : null, n, 20)
        }, T.toFormat = function (e, n) {
            var t = _(this, null != e && H(e, 0, E, 21) ? ~~e + this.e + 1 : null, n, 21);
            if (this.c) {
                var r, i = t.split("."), o = +X.groupSize, u = +X.secondaryGroupSize, s = X.groupSeparator, f = i[0],
                    l = i[1], c = this.s < 0, a = c ? f.slice(1) : f, h = a.length;
                if (u && (r = o, o = u, u = r, h -= r), o > 0 && h > 0) {
                    for (r = h % o || o, f = a.substr(0, r); h > r; r += o) f += s + a.substr(r, o);
                    u > 0 && (f += s + a.slice(r)), c && (f = "-" + f)
                }
                t = l ? f + X.decimalSeparator + ((u = +X.fractionGroupSize) ? l.replace(new RegExp("\\d{" + u + "}\\B", "g"), "$&" + X.fractionGroupSeparator) : l) : f
            }
            return t
        }, T.toFraction = function (e) {
            var n, t, i, o, u, s, f, l, c, h = j, g = this, p = g.c, d = new a(q), m = t = new a(q), w = f = new a(q);
            if (null != e && (j = !1, s = new a(e), j = h, (!(h = s.isInt()) || s.lt(q)) && (j && L(22, "max denominator " + (h ? "out of range" : "not an integer"), e), e = !h && s.c && U(s, s.e + 1, 1).gte(q) ? s : null)), !p) return g.toString();
            for (c = r(p), o = d.e = c.length - g.e - 1, d.c[0] = R[(u = o % y) < 0 ? y + u : u], e = !e || s.cmp(d) > 0 ? o > 0 ? d : m : s, u = z, z = 1 / 0, s = new a(c), f.c[0] = 0; l = C(s, d, 0, 1), i = t.plus(l.times(w)), 1 != i.cmp(e);) t = w, w = i, m = f.plus(l.times(i = m)), f = i, d = s.minus(l.times(i = d)), s = i;
            return i = C(e.minus(t), w, 0, 1), f = f.plus(i.times(m)), t = t.plus(i.times(w)), f.s = m.s = g.s, o *= 2, n = C(m, w, o, k).minus(g).abs().cmp(C(f, t, o, k).minus(g).abs()) < 1 ? [m.toString(), w.toString()] : [f.toString(), t.toString()], z = u, n
        }, T.toNumber = function () {
            var e = this;
            return +e || (e.s ? 0 * e.s : 0 / 0)
        }, T.toPower = T.pow = function (e) {
            var n, t, r = m(0 > e ? -e : +e), i = this;
            if (!H(e, -S, S, 23, "exponent") && (!isFinite(e) || r > S && (e /= 0) || parseFloat(e) != e && !(e = 0 / 0))) return new a(Math.pow(+i, e));
            for (n = J ? d(J / y + 2) : 0, t = new a(q); ;) {
                if (r % 2) {
                    if (t = t.times(i), !t.c) break;
                    n && t.c.length > n && (t.c.length = n)
                }
                if (r = m(r / 2), !r) break;
                i = i.times(i), n && i.c && i.c.length > n && (i.c.length = n)
            }
            return 0 > e && (t = q.div(t)), n ? U(t, J, k) : t
        }, T.toPrecision = function (e, n) {
            return _(this, null != e && H(e, 1, E, 24, "precision") ? 0 | e : null, n, 24)
        }, T.toString = function (e) {
            var n, t = this, i = t.s, o = t.e;
            return null === o ? i ? (n = "Infinity", 0 > i && (n = "-" + n)) : n = "NaN" : (n = r(t.c), n = null != e && H(e, 2, 64, 25, "base") ? D(l(n, o), 0 | e, 10, i) : B >= o || o >= $ ? f(n, o) : l(n, o), 0 > i && t.c[0] && (n = "-" + n)), n
        }, T.truncated = T.trunc = function () {
            return U(new a(this), this.e + 1, 1)
        }, T.valueOf = T.toJSON = function () {
            return this.toString()
        }, null != e && a.config(e), a
    }

    function t(e) {
        var n = 0 | e;
        return e > 0 || e === n ? n : n - 1
    }

    function r(e) {
        for (var n, t, r = 1, i = e.length, o = e[0] + ""; i > r;) {
            for (n = e[r++] + "", t = y - n.length; t--; n = "0" + n) ;
            o += n
        }
        for (i = o.length; 48 === o.charCodeAt(--i);) ;
        return o.slice(0, i + 1 || 1)
    }

    function i(e, n) {
        var t, r, i = e.c, o = n.c, u = e.s, s = n.s, f = e.e, l = n.e;
        if (!u || !s) return null;
        if (t = i && !i[0], r = o && !o[0], t || r) return t ? r ? 0 : -s : u;
        if (u != s) return u;
        if (t = 0 > u, r = f == l, !i || !o) return r ? 0 : !i ^ t ? 1 : -1;
        if (!r) return f > l ^ t ? 1 : -1;
        for (s = (f = i.length) < (l = o.length) ? f : l, u = 0; s > u; u++) if (i[u] != o[u]) return i[u] > o[u] ^ t ? 1 : -1;
        return f == l ? 0 : f > l ^ t ? 1 : -1
    }

    function o(e, n, t) {
        return (e = c(e)) >= n && t >= e
    }

    function u(e) {
        return "[object Array]" == Object.prototype.toString.call(e)
    }

    function s(e, n, t) {
        for (var r, i, o = [0], u = 0, s = e.length; s > u;) {
            for (i = o.length; i--; o[i] *= n) ;
            for (o[r = 0] += O.indexOf(e.charAt(u++)); r < o.length; r++) o[r] > t - 1 && (null == o[r + 1] && (o[r + 1] = 0), o[r + 1] += o[r] / t | 0, o[r] %= t)
        }
        return o.reverse()
    }

    function f(e, n) {
        return (e.length > 1 ? e.charAt(0) + "." + e.slice(1) : e) + (0 > n ? "e" : "e+") + n
    }

    function l(e, n) {
        var t, r;
        if (0 > n) {
            for (r = "0."; ++n; r += "0") ;
            e = r + e
        } else if (t = e.length, ++n > t) {
            for (r = "0", n -= t; --n; r += "0") ;
            e += r
        } else t > n && (e = e.slice(0, n) + "." + e.slice(n));
        return e
    }

    function c(e) {
        return e = parseFloat(e), 0 > e ? d(e) : m(e)
    }

    var a, h, g, p = /^-?(\d+(\.\d*)?|\.\d+)(e[+-]?\d+)?$/i, d = Math.ceil, m = Math.floor,
        w = " not a boolean or binary digit", v = "rounding mode",
        b = "number type has more than 15 significant digits",
        O = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ$_", N = 1e14, y = 14, S = 9007199254740991,
        R = [1, 10, 100, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13], A = 1e7, E = 1e9;
    if (a = n(), "function" == typeof define && define.amd) define(function () {
        return a
    }); else if ("undefined" != typeof module && module.exports) {
        if (module.exports = a, !h) try {
            h = require("crypto")
        } catch (D) {
        }
    } else e.BigNumber = a
}(this);

