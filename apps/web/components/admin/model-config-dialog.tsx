"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, Loader2 } from "lucide-react";
import { apiFetch } from "@/lib/api";

interface ModelConfigField {
  name: string;
  type: string;
  default: string | number | boolean;
  description: string;
  options?: string[];
  min?: number;
  max?: number;
  required: boolean;
}

interface ModelConfigSchema {
  model_id: string;
  display_name: string;
  fields: ModelConfigField[];
}

interface ModelConfig {
  model_id: string;
  config: Record<string, unknown>;
}

interface ModelConfigDialogProps {
  modelId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function ModelConfigDialog({
  modelId,
  open,
  onOpenChange,
  onSuccess,
}: ModelConfigDialogProps) {
  const [schema, setSchema] = useState<ModelConfigSchema | null>(null);
  const [config, setConfig] = useState<Record<string, unknown>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSchemaAndConfig = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch schema and current config in parallel
      const [schemaData, configData] = await Promise.all([
        apiFetch<ModelConfigSchema>(`/v1/admin/models/${encodeURIComponent(modelId)}/config/schema`),
        apiFetch<ModelConfig>(`/v1/admin/models/${encodeURIComponent(modelId)}/config`).catch(() => null),
      ]);

      setSchema(schemaData);
      
      // Set initial config values from current config or defaults
      const initialConfig: Record<string, unknown> = {};
      schemaData.fields.forEach((field) => {
        if (configData?.config && field.name in configData.config) {
          initialConfig[field.name] = configData.config[field.name];
        } else {
          initialConfig[field.name] = field.default;
        }
      });
      setConfig(initialConfig);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load configuration");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && modelId) {
      fetchSchemaAndConfig();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, modelId]);

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);

      await apiFetch(`/v1/admin/models/${encodeURIComponent(modelId)}/config`, {
        method: "PUT",
        body: JSON.stringify(config),
      });

      onSuccess();
      onOpenChange(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save configuration");
    } finally {
      setSaving(false);
    }
  };

  const updateConfigValue = (fieldName: string, value: unknown) => {
    setConfig((prev) => ({
      ...prev,
      [fieldName]: value,
    }));
  };

  const renderField = (field: ModelConfigField) => {
    const value = config[field.name];

    switch (field.type) {
      case "bool":
        return (
          <div key={field.name} className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor={field.name}>{formatFieldName(field.name)}</Label>
                <p className="text-sm text-gray-500 dark:text-gray-400">{field.description}</p>
              </div>
              <Switch
                id={field.name}
                checked={value as boolean}
                onCheckedChange={(checked: boolean) => updateConfigValue(field.name, checked)}
              />
            </div>
          </div>
        );

      case "string":
        if (field.options && field.options.length > 0) {
          return (
            <div key={field.name} className="space-y-2">
              <Label htmlFor={field.name}>{formatFieldName(field.name)}</Label>
              <Select
                value={value as string}
                onValueChange={(newValue: string) => updateConfigValue(field.name, newValue)}
              >
                <SelectTrigger id={field.name}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {field.options.map((option) => (
                    <SelectItem key={option} value={option}>
                      {option}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-sm text-gray-500 dark:text-gray-400">{field.description}</p>
            </div>
          );
        }
        return (
          <div key={field.name} className="space-y-2">
            <Label htmlFor={field.name}>{formatFieldName(field.name)}</Label>
            <Input
              id={field.name}
              type="text"
              value={value as string}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateConfigValue(field.name, e.target.value)}
            />
            <p className="text-sm text-gray-500 dark:text-gray-400">{field.description}</p>
          </div>
        );

      case "int":
        return (
          <div key={field.name} className="space-y-2">
            <Label htmlFor={field.name}>{formatFieldName(field.name)}</Label>
            <Input
              id={field.name}
              type="number"
              min={field.min}
              max={field.max}
              value={value as number}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateConfigValue(field.name, parseInt(e.target.value, 10))}
            />
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {field.description}
              {field.min !== undefined && field.max !== undefined && (
                <span className="ml-1">({field.min} - {field.max})</span>
              )}
            </p>
          </div>
        );

      case "float":
        return (
          <div key={field.name} className="space-y-2">
            <Label htmlFor={field.name}>{formatFieldName(field.name)}</Label>
            <Input
              id={field.name}
              type="number"
              step="0.1"
              min={field.min}
              max={field.max}
              value={value as number}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateConfigValue(field.name, parseFloat(e.target.value))}
            />
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {field.description}
              {field.min !== undefined && field.max !== undefined && (
                <span className="ml-1">({field.min} - {field.max})</span>
              )}
            </p>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Configure {schema?.display_name || "Model"}</DialogTitle>
          <DialogDescription>
            Adjust configuration parameters for this AI model. Changes take effect immediately for new jobs.
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
          </div>
        ) : error ? (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : schema ? (
          <div className="space-y-6 py-4">
            {schema.fields.map((field) => renderField(field))}
          </div>
        ) : null}

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={loading || saving}
          >
            {saving ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Save Configuration"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function formatFieldName(name: string): string {
  return name
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
