import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { PromoCodeFormData } from "./promo-code-form";

interface PromoCodeLimitsFieldsProps {
  formData: PromoCodeFormData;
  onFormChange: (data: PromoCodeFormData) => void;
}

export function PromoCodeLimitsFields({
  formData,
  onFormChange,
}: PromoCodeLimitsFieldsProps) {
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="duration_months">订阅时长（月） *</Label>
        <Input
          id="duration_months"
          type="number"
          min={1}
          max={24}
          value={formData.duration_months}
          onChange={(e) =>
            onFormChange({
              ...formData,
              duration_months: parseInt(e.target.value) || 1,
            })
          }
          required
        />
        <p className="text-xs text-muted-foreground">
          订阅将被延长的月数
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="max_uses">总使用次数上限</Label>
          <Input
            id="max_uses"
            type="number"
            min={1}
            placeholder="不限"
            value={formData.max_uses}
            onChange={(e) =>
              onFormChange({ ...formData, max_uses: e.target.value })
            }
          />
          <p className="text-xs text-muted-foreground">
            留空表示不限次数
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="max_uses_per_org">每个组织使用次数上限</Label>
          <Input
            id="max_uses_per_org"
            type="number"
            min={1}
            value={formData.max_uses_per_org}
            onChange={(e) =>
              onFormChange({
                ...formData,
                max_uses_per_org: parseInt(e.target.value) || 1,
              })
            }
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="expires_at">过期时间</Label>
        <Input
          id="expires_at"
          type="datetime-local"
          value={formData.expires_at}
          onChange={(e) =>
            onFormChange({ ...formData, expires_at: e.target.value })
          }
        />
        <p className="text-xs text-muted-foreground">
          留空表示永不过期
        </p>
      </div>
    </>
  );
}
