<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT%20with%20Attribution-blue?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Zero%20Dependencies-lightgrey?style=for-the-badge" alt="Zero Dependencies">
</p>

<h1 align="center">dhook</h1>

<p align="center">
  یک SDK آماده تولید، سازمانی و با عملکرد بالا برای وب‌هوک‌های دیسکورد در زبان Go<br>
  مسیریابی چند آدرس، محدودیت نرخ خودکار، صف مقاوم و سازنده embed روان
</p>

---

## ویژگی‌ها

- **مسیریابی چند آدرس** — ارسال همزمان پیام‌ها به چندین وب‌هوک
- **محدودیت نرخ خودکار** — رعایت محدودیت‌های نرخ دیسکورد با تلاش مجدد خودکار
- **پشتیبانی از Context** — تمام متدها `context.Context` را برای مدیریت تایم‌اوت و لغو دریافت می‌کنند
- **سازنده Embed روان** — API زنجیره‌ای برای ساخت embed‌های غنی
- **آپلود فایل** — ارسال فایل‌ها و پیوست‌ها از طریق multipart form data
- **صف در پس‌زمینه** — مجموعه کارگرهای همزمان برای ارسال با توان بالا
- **هوک‌های رویداد** — ثبت callback برای موفقیت، محدودیت نرخ و خطاها
- **بدون وابستگی** — کاملاً بر پایه کتابخانه استاندارد Go ساخته شده

## نصب

```bash
go get github.com/RivnZero/dhook
```

## شروع سریع

### ارسال پیام ساده

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/RivnZero/dhook"
)

func main() {
    client := dhook.New(
        "https://discord.com/api/webhooks/ID/TOKEN",
    )

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    msg := &dhook.Message{
        Content: "سلام از dhook!",
    }

    responses, err := client.Send(ctx, msg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("به %d وب‌هوک ارسال شد", len(responses))
}
```

### ارسال Embed غنی

```go
embed := dhook.NewEmbed().
    SetTitle("بیلد کامل شد").
    SetDescription("تمام تست‌ها با موفقیت اجرا شدند.").
    SetColor(0x57F287).
    AddField("وضعیت", "موفق", true).
    AddField("مدت زمان", "12 ثانیه", true).
    SetFooter("خط لوله CI", "").
    SetTimestamp(time.Now())

msg := &dhook.Message{
    Content: "**گزارش بیلد**",
    Embeds:  []*dhook.Embed{embed},
}

client.Send(ctx, msg)
```

### ارسال فایل

```go
file, _ := os.Open("report.pdf")
defer file.Close()

msg := &dhook.Message{
    Content: "گزارش ضمیمه شده است.",
}

client.SendFile(ctx, "report.pdf", file, msg)
```

### مسیریابی چند آدرس

```go
client := dhook.New(
    "https://discord.com/api/webhooks/ID_1/TOKEN_1",
    "https://discord.com/api/webhooks/ID_2/TOKEN_2",
    "https://discord.com/api/webhooks/ID_3/TOKEN_3",
)

msg := &dhook.Message{Content: "ارسال همزمان به تمام کانال‌ها!"}
responses, err := client.Send(ctx, msg)
```

### هوک‌های رویداد

```go
client.AddHook(dhook.EventSuccess, func(resp *dhook.Response) {
    log.Printf("پیام تحویل داده شد: %s", resp.ID)
})

client.AddHook(dhook.EventRateLimit, func(retryAfter time.Duration) {
    log.Printf("محدودیت نرخ - تلاش مجدد بعد از %v", retryAfter)
})

client.AddHook(dhook.EventError, func(err error) {
    log.Printf("خطا: %v", err)
})
```

### صف در پس‌زمینه

```go
queue := dhook.NewQueue(client, 10)
queue.Start(ctx)
defer queue.Stop()

for i := 0; i < 1000; i++ {
    queue.Add(&dhook.Message{
        Content: fmt.Sprintf("پیام شماره %d", i),
    })
}
```

## مرجع API

### کلاینت

| متد | توضیحات |
|-----|---------|
| `New(urls ...string) *Client` | ایجاد کلاینت جدید با یک یا چند آدرس وب‌هوک |
| `SetHTTPClient(client *http.Client)` | جایگزینی کلاینت HTTP پیش‌فرض |
| `Send(ctx, msg) ([]*Response, error)` | ارسال پیام به تمام وب‌هوک‌های پیکربندی شده |
| `SendFile(ctx, name, reader, msg) ([]*Response, error)` | ارسال فایل با پیام اختیاری |
| `SendFiles(ctx, msg, files...) ([]*Response, error)` | ارسال چند فایل با پیام |
| `Edit(ctx, messageID, msg) (*Response, error)` | ویرایش پیام ارسال شده |
| `Delete(ctx, messageID) error` | حذف پیام |
| `AddHook(event, callback)` | ثبت callback رویداد |

### سازنده Embed

| متد | توضیحات |
|-----|---------|
| `NewEmbed() *Embed` | ایجاد embed جدید |
| `SetTitle(string)` | تنظیم عنوان embed |
| `SetDescription(string)` | تنظیم توضیحات embed |
| `SetColor(int)` | تنظیم رنگ embed (عدد هگزادسیمال) |
| `SetURL(string)` | تنظیم URL embed |
| `SetTimestamp(time.Time)` | تنظیم زمان embed |
| `SetFooter(text, iconURL)` | تنظیم پاورقی embed |
| `SetImage(url)` | تنظیم تصویر embed |
| `SetThumbnail(url)` | تنظیم تصویر کوچک embed |
| `SetAuthor(name, url, iconURL)` | تنظیم نویسنده embed |
| `AddField(name, value, inline)` | افزودن فیلد به embed |

### صف

| متد | توضیحات |
|-----|---------|
| `NewQueue(client, workerCount)` | ایجاد صف جدید با N کارگر |
| `Start(ctx)` | شروع پردازش کارها |
| `Stop()` | توقف پردازش و تخلیه کارهای باقی‌مانده |
| `Add(msg)` | افزودن پیام به صف برای ارسال |
| `AddFunc(fn)` | افزودن تابع سفارشی برای اجرا |
| `Len()` | تعداد کارهای در انتظار |
| `Cap()` | حداکثر ظرفیت صف |

## ابزار خط فرمان

dhook یک ابزار خط فرمان مستقل ارائه می‌دهد که از مثال advanced ساخته شده است. از آن برای ارسال پیام مستقیماً از ترمینال استفاده کنید:

```bash
dhook \
  --urls "https://discord.com/api/webhooks/ID/TOKEN" \
  --content "سلام از خط فرمان!" \
  --embed-title "استقرار کامل شد" \
  --embed-desc "تمام سرویس‌ها فعال هستند." \
  --embed-color 0x57F287
```

### پارامترهای خط فرمان

| پارامتر | توضیحات | پیش‌فرض |
|---------|---------|---------|
| `--urls` | آدرس‌های وب‌هوک جدا شده با کاما | (الزامی) |
| `--content` | محتوای پیام | |
| `--username` | نمایش نام کاربری سفارشی | |
| `--avatar` | آدرس آواتار سفارشی | |
| `--file` | مسیر فایل پیوست | |
| `--filename` | نام فایل پیوست سفارشی | (از نام فایل) |
| `--embed-title` | عنوان embed | |
| `--embed-desc` | توضیحات embed | |
| `--embed-color` | رنگ embed به صورت هگزادسیمال | `0` |
| `--queue` | استفاده از کارگرهای صف پس‌زمینه | `false` |
| `--workers` | تعداد کارگرهای صف | `5` |
| `--timeout` | تایم‌اوت درخواست | `30s` |

## خط لوله استقرار و انتشار

### پیش‌انتشار محلی (deploy.bat)

اسکریپت `deploy.bat` دروازه‌بان محلی قبل از هر انتشار است. آن را اجرا کنید تا مطمئن شوید پروژه آماده تولید است:

```bat
deploy.bat
```

این اسکریپت مراحل زیر را انجام می‌دهد:

1. **`go vet ./...`** — تحلیل ایستای کد برای مشکلات رایج
2. **`go test -v ./...`** — اجرای تمام تست‌ها (در اولین خطا متوقف می‌شود)
3. **شبیه‌سازی کامپایل متقاطع** — کامپایل ابزار خط فرمان برای ۹ هدف:
   - `windows/386`، `windows/amd64`، `windows/arm64`
   - `linux/386`، `linux/amd64`، `linux/arm64`
   - `darwin/amd64`، `darwin/arm64`
4. **نتیجه** — اگر تمام بیلدها موفق باشند، دستورات تگ زدن ایمن را نمایش می‌دهد:

```
git tag vX.Y.Z
git push origin vX.Y.Z
```

### انتشار ابری (GoReleaser + GitHub Actions)

وقتی یک تگ نسخه معنایی پوش کنید، workflow فایل `release.yml` به‌طور خودکار:

1. روی هر تگ `v*` پوش شده فعال می‌شود
2. آخرین نسخه Go را راه‌اندازی می‌کند
3. GoReleaser را اجرا می‌کند که:
   - برای تمام اهداف OS/arch کامپایل متقاطع انجام می‌دهد (به جز `darwin/386`)
   - آرشیوهای `.zip` برای ویندوز و `.tar.gz` برای لینوکس/مک تولید می‌کند
   - تغییرات را دسته‌بندی شده بر اساس ویژگی‌ها و رفع باگ‌ها تولید می‌کند
   - یک انتشار گیت‌هاب با تمام آرشیوها و checksums ایجاد می‌کند

### فرآیند کامل انتشار

```bash
# ۱. اجرای دروازه‌بان محلی
deploy.bat

# ۲. تگ زدن و پوش کردن
git tag v1.0.0
git push origin v1.0.0

# ۳. GitHub Actions بقیه کار را به‌طور خودکار انجام می‌دهد
```

## نحوه کار

### محدودیت نرخ

dhook هدرهای `X-Rate-Limit-Remaining` و `Retry-After` دیسکورد را در هر پاسخ می‌خواند. وقتی به محدودیت می‌رسید، مدیریت‌کننده نرخ به‌طور خودکار درخواست‌های بعدی را تا پایان دوره انتظار مسدود می‌کند. بدون از دست دادن پیام، بدون منطق تلاش مجدد دستی.

### پشتیبان اتصال نمایی

در صورت خطاهای سرور 5xx، dhook با پشتیبان اتصال نمایی از ۱ ثانیه شروع می‌کند، تا ۳۰ ثانیه دو برابر می‌شود و حداکثر ۵ بار تلاش مجدد برای هر درخواست انجام می‌دهد.

### پخش همزمان

وقتی چندین آدرس وب‌هوک پیکربندی شده باشد، `Send` درخواست‌ها را به‌طور همزمان با goroutine به تمام آدرس‌ها ارسال می‌کند. هر آدرس به‌طور مستقل مدیریت نرخ می‌شود.

## مجوز

مجوز MIT با الزام انتساب. جزئیات را در [LICENSE](LICENSE) ببینید.

**نویسنده: محمد کیان استادمحمودی**
